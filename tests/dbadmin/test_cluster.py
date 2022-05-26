"""
Test suite to validate clusters endpoints operations
"""
import copy
import random
import urllib
from common import get_config, API
from cluster import Cluster
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_cluster_get_all():
    """
    Test get all clusters
    """
    cluster_api = Cluster(config, session)
    cluster_api.get_all()


def test_cluster_crud_lifecycle():
    """
    Test create, read, update, delete one cluster
    """
    cluster_api = Cluster(config, session)
    created_cluster = cluster_api.create_positive()

    # Creating same, now existing cluster, must not be possible
    cluster_api.create_negative(created_cluster)

    # Read existing cluster by name
    retrieved_cluster = cluster_api.get_positive(created_cluster['name'])
    cluster_api.assert_compare(retrieved_cluster, created_cluster)

    # Check if created cluster shows up in global cluster list
    assert(True for l in cluster_api.get_all()['cluster'] if l['name'] == created_cluster['name'])

    # Update to new email address and attributes
    updated_cluster = copy.deepcopy(created_cluster)
    random_int = random.randint(0,99999)
    updated_cluster['displayName'] = f'Testsuite Cluster Name {random_int}'
    updated_cluster['attributes'] = [
        {
            "name": "IdleTimeout",
            "value": "42ms"
        },
        {
            "name": "SNIHostName",
            "value": "eb"
        }
    ]
    cluster_api.update_positive(created_cluster['name'], updated_cluster)

    # Check if updated cluster shows up in global cluster list
    assert(True for l in cluster_api.get_all()['cluster'] if l['name'] == created_cluster['name'])

    # Delete just created cluster
    deleted_cluster = cluster_api.delete_positive(updated_cluster['name'])
    cluster_api.assert_compare(deleted_cluster, updated_cluster)

    # Try to delete cluster once more, must not exist anymore
    cluster_api.delete_negative(updated_cluster['name'])
