"""
Developer module does all REST API operations on developer endpoint
"""
import random
from common import assert_valid_schema, assert_status_code, \
    assert_content_type_json, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Developer:
    """
    Developer does all REST API operations on developer endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.developer_url = config['api_url'] + '/developers'
        self.schemas = {
            'developer': load_json_schema('developer.json'),
            'developers': load_json_schema('developers.json'),
            'developers_email': load_json_schema('developers-email-addresses.json'),
            'error': load_json_schema('error.json'),
        }


    def generate_email_address(self, number):
        """
        Returns email address of test developer based upon provided number
        """
        return f"testsuite-dev{number}@test.com"


    def create_new(self, new_developer=None):
        """
        Create new developer to be used as test subject. If none provided generate random developer
        """
        if new_developer is None:
            random_int = random.randint(0,99999)
            new_developer = {
                "email" : self.generate_email_address(random_int),
                "firstName" : f"first{random_int}",
                "lastName" : f"last{random_int}",
                "userName" : f"username{random_int}",
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.developer_url, json=new_developer)
        assert_status_code(response, HTTP_CREATED)
        assert_content_type_json(response)

        # Check if just created developer matches with what we requested
        created_developer = response.json()
        assert_valid_schema(created_developer, self.schemas['developer'])
        self.assert_compare(created_developer, new_developer)

        return created_developer


    def create_existing(self, developer):
        """
        Attempt to create new developer which already exists, should fail
        """
        response = self.session.post(self.developer_url, json=developer)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['error'])


    def get_all(self):
        """
        Get all developers email addresses
        """
        response = self.session.get(self.developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['developers_email'])

        return response.json()


    def get_all_detailed(self):
        """
        Get all developers with full details
        """
        response = self.session.get(self.developer_url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['developers'])
        # TODO testing of paginating response


    def get_existing(self, developer_email):
        """
        Get existing developer
        """
        developer_url = self.developer_url + '/' + developer_email
        response = self.session.get(developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_developer = response.json()
        assert_valid_schema(retrieved_developer, self.schemas['developer'])

        return retrieved_developer


    def update_existing(self, developer_email, developer):
        """
        Update existing developer
        """
        developer_url = self.developer_url + '/' + developer_email
        response = self.session.post(developer_url, json=developer)
        print(response.text)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_developer = response.json()
        assert_valid_schema(updated_developer, self.schemas['developer'])

        return updated_developer


    def delete_existing(self, developer_email):
        """
        Delete existing developer
        """
        developer_url = self.developer_url + '/' + developer_email
        response = self.session.delete(developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        deleted_developer = response.json()
        assert_valid_schema(deleted_developer, self.schemas['developer'])

        return deleted_developer


    def delete_nonexisting(self, developer_email):
        """
        Delete non-existing developer, which should fail
        """
        developer_url = self.developer_url + '/' + developer_email
        response = self.session.delete(developer_url)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete_all_test_developer(self):
        """
        Delete all developers created by test suite
        """
        for i in range(self.config['entity_count']):
            email = self.generate_email_address(i)
            developer_url = self.developer_url + '/' + email
            self.session.delete(developer_url)


    def assert_compare(self, developer_a, developer_b):
        """
        Compares minimum required fields that can be set of two developers
        """
        assert developer_a['email'] == developer_b['email']
        assert developer_a['firstName'] == developer_b['firstName']
        assert developer_a['lastName'] == developer_b['lastName']
        assert developer_a['attributes'] == developer_b['attributes']
