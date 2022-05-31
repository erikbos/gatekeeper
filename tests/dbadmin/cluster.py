"""
Cluster module does all REST API operations on cluster endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Cluster:
    """
    Cluster does all REST API operations on cluster endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/v1/clusters'


    def generate_cluster_name(self, number):
        """
        Returns name of test cluster based upon provided number
        """
        return f"testsuite-cluster{number}"


    def _create(self, success_expected, new_cluster=None):
        """
        Create new cluster to be used as test subject.
        If none provided generate cluster with random name.
        """

        if new_cluster is None:
            random_int = random.randint(0,99999)
            new_cluster = {
                "name" : self.generate_cluster_name(random_int),
                "displayName": f'TestSuite Cluster {random_int}',
                "attributes": [
                    {
                        "name": "SNIHostName",
                        "value": f"testsuite{random_int}.example.com"
                    }
                ],
            }

        response = self.session.post(self.url, json=new_cluster)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)

            # Check if just created cluster matches with what we requested
            created_cluster = response.json()
            self.assert_compare(created_cluster, new_cluster)
            return created_cluster

        assert_status_code(response, HTTP_BAD_REQUEST)
        return response.json()


    def create_positive(self, new_cluster=None):
        """
        Create new cluster, if none provided generate cluster with random data.
        """
        return self._create(True, new_cluster)


    def create_negative(self, cluster):
        """
        Attempt to create cluster which already exists, should fail
        """
        return self._create(False, cluster)


    def get_all(self):
        """
        Get all clusters
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)

        return response.json()


    def get_all_detailed(self):
        """
        Get all clusters with full details
        """
        response = self.session.get(self.url + '?expand=true')
        assert_status_code(response, HTTP_OK)


    def get_positive(self, cluster):
        """
        Get existing cluster
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(cluster))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, cluster, updated_cluster):
        """
        Update existing cluster
        """
        response = self.session.post(self.url + '/' + urllib.parse.quote(cluster), json=updated_cluster)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _delete(self, cluster_name, expected_success):
        """
        Delete existing cluster
        """
        cluster_url = self.url + '/' + urllib.parse.quote(cluster_name)
        response = self.session.delete(cluster_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)

        return response.json()


    def delete_positive(self, cluster_name):
        """
        Delete existing cluster
        """
        return self._delete(cluster_name, True)


    def delete_negative(self, cluster_name):
        """
        Attempt to delete non-existing cluster, should fail
        """
        return self._delete(cluster_name, False)


    def assert_compare(self, cluster_a, cluster_b):
        """
        Compares minimum required fields that can be set of two clusters
        """
        assert cluster_a['name'] == cluster_b['name']
        assert (cluster_a['attributes'].sort(key=lambda x: x['name'])
            == cluster_b['attributes'].sort(key=lambda x: x['name']))
