"""
Test suite to validate organizations endpoints operations
"""
import copy
import random
import urllib
from common import get_config, API
from organization import Organization
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_organization_get_all():
    """
    Test get all organizations
    """
    organization_api = Organization(config, session)
    organization_api.get_all()


def test_organization_crud_lifecycle():
    """
    Test create, read, update, delete one organization
    """
    organization_api = Organization(config, session)
    created_organization = organization_api.create_positive()

    # Creating same, now existing organization, must not be possible
    organization_api.create_negative(created_organization)

    # Read existing organization by name
    retrieved_organization = organization_api.get_positive(created_organization['name'])
    organization_api.assert_compare(retrieved_organization, created_organization)

    # Check if created organization shows up in global organization list
    assert(True for l in organization_api.get_all()['organization'] if l['name'] == created_organization['name'])

    # Update to new email address and attributes
    updated_organization = copy.deepcopy(created_organization)
    random_int = random.randint(0,99999)
    updated_organization['displayName'] = f'Testsuite Organization Name {random_int}'
    updated_organization['attributes'] = [
        {
            "name": "HostHeader",
            "value": "42.example.com"
        }
    ]
    organization_api.update_positive(created_organization['name'], updated_organization)

    # Check if updated organization shows up in global organization list
    assert(True for l in organization_api.get_all()['organization'] if l['name'] == created_organization['name'])

    # Delete just created organization
    deleted_organization = organization_api.delete_positive(updated_organization['name'])
    organization_api.assert_compare(deleted_organization, updated_organization)

    # Try to delete organization once more, must not exist anymore
    organization_api.delete_negative(updated_organization['name'])
