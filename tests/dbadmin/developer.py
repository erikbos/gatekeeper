import random
from common import assert_valid_schema, assert_status_code, assert_content_type_json, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Developer:
    """
    Developer does all REST API functions on developer endpoint
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

        response = self.session.post(self.developer_url, json=new_developer)
        assert_status_code(response, HTTP_CREATED)
        assert_content_type_json(response)

        # Check if just created developer matches with what we requested
        created_developer = response.json()
        assert_valid_schema(created_developer, self.schemas['developer'])
        self.assert_compare_developer(created_developer, new_developer)

        return created_developer


    def create_existing_developer(self, developer):
        """
        Attempt to create new developer which already exists, should fail
        """
        response = self.session.post(self.developer_url, json=developer)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas)


    def get_all(self):
        """
        Get all developers email addresses
        """
        response = self.session.get(self.developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['developers_email'])


    def get_all_detailed(self):
        """
        Get all developers with full details
        """
        response = self.session.get(self.developer_url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['developers'])
        # TODO (erikbos) testing of paginating


    def get_existing(self, id):
        """
        Get existing developer
        """
        developer_url = self.developer_url + '/' + id
        response = self.session.get(developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_developer = response.json()
        assert_valid_schema(retrieved_developer, self.schemas['developer'])

        return retrieved_developer


    def update_existing(self, developer):
        """
        Update existing developer
        """
        developer_url = self.developer_url + '/' + developer['developerId']
        response = self.session.post(developer_url, json=developer)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_developer = response.json()
        assert_valid_schema(updated_developer, self.schemas['developer'])

        return updated_developer


    def delete_existing(self, id):
        """
        Delete existing developer
        """
        developer_url = self.developer_url + '/' + id
        response = self.session.delete(developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        deleted_developer = response.json()
        assert_valid_schema(deleted_developer, self.schemas['developer'])

        return deleted_developer


    def delete_nonexisting(self, id):
        """
        Delete non-existing developer, which should fail
        """
        developer_url = self.developer_url + '/' + id
        response = self.session.delete(developer_url)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete_all_test_developer(self):
        for i in range(self.config['entity_count']):
            new_developer = {
                "email" : self.generate_email_address(i),
            }
            developer_url = self.developer_url + '/' + new_developer['email']
            self.session.delete(developer_url)


    def assert_compare_developer(self, a, b):
        """
        Compares minimum required fields that can be set of two developers
        """
        assert a['email'] == b['email']
        assert a['firstName'] == b['firstName']
        assert a['lastName'] == b['lastName']
