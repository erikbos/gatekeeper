import requests
import random
from common import assert_valid_schema, assert_status_code, assert_content_type_json, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST
from attribute import run_attribute_tests


class Developer:
    """
    Developer does all REST API functions on developer endpoint
    """
    def __init__(self, config):
        self.config = config
        self.developer_url = config['api_url'] + '/developers'
        self.schema_developer = load_json_schema('developer.json')
        self.schema_developers = load_json_schema('developers.json')
        self.schema_developers_email = load_json_schema('developers-email-addresses.json')
        self.schema_error = load_json_schema('error.json')

    def _get_auth(self):
        """Returns HTTP auth parameters"""

        if 'api_username' in self.config and 'api_password' in self.config:
            return (self.config['api_username'], self.config['api_password'])
        else:
            return None


    def generate_email_address(self, number):
        """
        Returns email address of test developer based upon provided number
        """
        return "testuser{}@test.com".format(number)


    def create_new(self, new_developer=None):
        """
        Create new developer to be used as test subject. If none provided generate random developer data
        """
        if new_developer == None:
            r = random.randint(0,99999)
            new_developer = {
                "email" : self.generate_email_address(r),
                "firstName" : f"first{r}",
                "lastName" : f"last{r}",
                "userName" : f"username{r}",
                "attributes" : [ ],
            }

        response = requests.post(self.developer_url, auth=self._get_auth(), headers=self.config['request_headers'], json=new_developer)
        assert_status_code(response, HTTP_CREATED)
        assert_content_type_json(response)

        # Check if just created developer matches with what we requested
        created_developer = response.json()
        assert_valid_schema(created_developer, self.schema_developer)
        self.assert_compare_developer(created_developer, new_developer)

        return created_developer


    def create_existing_developer(self, developer):
        """
        Attempt to create new developer which already exists, should fail
        """
        response = requests.post(self.developer_url, auth=self._get_auth(), headers=self.config['request_headers'], json=developer)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schema_error)


    def get_all(self):
        """
        Get all developers email addresses
        """
        response = requests.get(self.developer_url, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schema_developers_email)


    def get_all_detailed(self):
        """
        Get all developers with full details
        """
        response = requests.get(self.developer_url + '?expand=true', auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schema_developers)
        # TODO (erikbos) testing of paginating


    def get_existing(self, id):
        """
        Get existing developer
        """
        developer_url = self.developer_url + '/' + id
        response = requests.get(developer_url, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_developer = response.json()
        assert_valid_schema(retrieved_developer, self.schema_developer)

        return retrieved_developer


    def update_existing(self, developer):
        """
        Update existing developer
        """
        developer_url = self.developer_url + '/' + developer['developerId']
        response = requests.post(developer_url, auth=self._get_auth(), headers=self.config['request_headers'], json=developer)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_developer = response.json()
        assert_valid_schema(updated_developer, self.schema_developer)

        return updated_developer


    def delete_existing(self, id):
        """
        Delete existing developer
        """
        developer_url = self.developer_url + '/' + id
        response = requests.delete(developer_url, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        deleted_developer = response.json()
        assert_valid_schema(deleted_developer, self.schema_developer)

        return deleted_developer


    def delete_nonexisting(self, id):
        """
        Delete non-existing developer, which should fail
        """
        developer_url = self.developer_url + '/' + id
        response = requests.delete(developer_url, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete_all_test_developer(self):
        for i in range(self.config['entity_count']):
            new_developer = {
                "email" : self.generate_email_address(i),
            }
            developer_url = self.developer_url + '/' + new_developer['email']
            requests.delete(developer_url, auth=self._get_auth())


    def assert_compare_developer(self, a, b):
        """
        Compares minimum required fields that can be set of two developers
        """
        assert a['email'] == b['email']
        assert a['firstName'] == b['firstName']
        assert a['lastName'] == b['lastName']
