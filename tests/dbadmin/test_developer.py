import requests
import random
from common import assert_valid_schema, assert_status_code, assert_content_type_json, get_config, default_headers
from httperror import *
from attribute import run_attributes_test

config = get_config()

# Number of developers to create while CRUDing
developer_crud_test_count = 10

class Developer:
    """
    Developer does all REST API functions on developer endpoint
    """
    def __init__(self, config):
        self.config = config
        self.developer_url = config['api_url'] + '/developers'


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

        response = requests.post(self.developer_url, auth=self._get_auth(), headers=default_headers, json=new_developer)
        assert_status_code(response, HTTP_CREATED)
        assert_content_type_json(response)

        # Check if just created developer matches with what we requested
        created_developer = response.json()
        assert_valid_schema(created_developer, 'developer.json')
        self.assert_compare_developer(created_developer, new_developer)

        return created_developer


    def create_existing_developer(self, developer):
        """
        Attempt to create new developer which already exists, should fail
        """
        response = requests.post(self.developer_url, auth=self._get_auth(), headers=default_headers, json=developer)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), 'error.json')


    def get_all(self):
        """
        Get all developers email addresses
        """
        response = requests.get(self.developer_url, auth=self._get_auth(), headers=default_headers)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), 'developers-email-addresses.json')


    def get_all_detailed(self):
        """
        Get all developers with full details
        """
        response = requests.get(self.developer_url + '?expand=true', auth=self._get_auth(), headers=default_headers)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), 'developers.json')
        # TODO (erikbos) testing of paginating


    def get_existing(self, id):
        """
        Get existing developer
        """
        developer_url = self.developer_url + '/' + id
        response = requests.get(developer_url, auth=self._get_auth(), headers=default_headers)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_developer = response.json()
        assert_valid_schema(retrieved_developer, 'developer.json')

        return retrieved_developer


    def update_existing(self, developer):
        """
        Update existing developer
        """
        developer_url = self.developer_url + '/' + developer['developerId']
        response = requests.post(developer_url, auth=self._get_auth(), headers=default_headers, json=developer)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_developer = response.json()
        assert_valid_schema(updated_developer, 'developer.json')

        return updated_developer


    def delete_existing(self, id):
        """
        Delete existing developer
        """
        developer_url = self.developer_url + '/' + id
        response = requests.delete(developer_url, auth=self._get_auth(), headers=default_headers)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        deleted_developer = response.json()
        assert_valid_schema(deleted_developer, 'developer.json')

        return deleted_developer


    def delete_nonexisting(self, id):
        """
        Delete non-existing developer, which should fail
        """
        developer_url = self.developer_url + '/' + id
        response = requests.delete(developer_url, auth=self._get_auth(), headers=default_headers)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete_all_test_developer(self):
        for i in range(developer_crud_test_count):
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


def test_get_all_developers_list():
    """
    Test get all developers, returns array of email addresses of developers
    """
    developerAPI = Developer(config)
    developerAPI.get_all()


def test_get_all_developers_list_detailed():
    """
    Test get all developers, returns array with zero or more developers
    """
    developerAPI = Developer(config)
    developerAPI.get_all_detailed()


def test_crud_lifecycle_one_developer():
    """
    Test create, read, update, delete one developer
    """
    developerAPI = Developer(config)

    created_developer = developerAPI.create_new()

    # Creating same, now existing developer, must not be possible
    developerAPI.create_existing_developer(created_developer)

    # Read existing developer by email
    retrieved_developer = developerAPI.get_existing(created_developer['email'])
    developerAPI.assert_compare_developer(retrieved_developer, created_developer)

    # Read existing developer by developerid
    retrieved_developer = developerAPI.get_existing(created_developer['developerId'])
    developerAPI.assert_compare_developer(retrieved_developer, created_developer)

    # Update email address while preserving developerID
    updated_developer = created_developer
    updated_developer['email'] = 'newemailaddress@test.com'
    updated_developer['attributes'] = [
              {
                   "name" : "Status"
              },
              {
                   "name" : "Shoesize",
                   "value" : "42"
              }
         ]
    developerAPI.update_existing(updated_developer)

    # Read updated developer by developerid
    retrieved_developer = developerAPI.get_existing(updated_developer['email'])
    developerAPI.assert_compare_developer(retrieved_developer, updated_developer)

    # Delete just created developer
    deleted_developer = developerAPI.delete_existing(updated_developer['email'])
    developerAPI.assert_compare_developer(deleted_developer, updated_developer)

    # Try to delete developer once more, must not exist anymore
    developerAPI.delete_nonexisting(updated_developer['email'])


def test_crud_lifecycle_ten_developers():
    """
    Test create, read, update, delete ten developers
    """
    developerAPI = Developer(config)

    # Create developers
    created_developers = []
    for i in range(developer_crud_test_count):
        created_developers.append(developerAPI.create_new())

    # Creating them once more must not be possible
    random.shuffle(created_developers)
    for i in range(developer_crud_test_count):
        developerAPI.create_existing_developer(created_developers[i])

    # Read created developers in random order
    random.shuffle(created_developers)
    for i in range(developer_crud_test_count):
        retrieved_developer = developerAPI.get_existing(created_developers[i]['developerId'])
        developerAPI.assert_compare_developer(retrieved_developer, created_developers[i])

    # Delete each created developer
    random.shuffle(created_developers)
    for i in range(developer_crud_test_count):
        deleted_developer = developerAPI.delete_existing(created_developers[i]['developerId'])
        # Check if just deleted developer matches with as created developer
        developerAPI.assert_compare_developer(deleted_developer, created_developers[i])

        # Try to delete developer once more, must not exist anymore
        developerAPI.delete_nonexisting(created_developers[i]['developerId'])


def test_change_developer_status():
    """
    Test changing status of developer
    """
    developerAPI = Developer(config)

    test_developer = developerAPI.create_new()

    # Change status to active
    url = developerAPI.developer_url + '/' + test_developer['email'] + '?action=active'
    response = requests.post(url, auth=developerAPI._get_auth(), headers=default_headers)
    assert_status_code(response, 204)
    assert response.content == b''

    assert developerAPI.get_existing(test_developer['email'])['status'] == 'active'

    # Change status to inactive
    url = developerAPI.developer_url + '/' + test_developer['email'] + '?action=inactive'
    response = requests.post(url, auth=developerAPI._get_auth(), headers=default_headers)
    assert_status_code(response, 204)
    assert response.content == b''

    assert developerAPI.get_existing(test_developer['email'])['status'] == 'inactive'

    # Change status to unknown value is unsupported
    url = developerAPI.developer_url + '/' + test_developer['email'] + '?action=unknown'
    response = requests.post(url, auth=developerAPI._get_auth(), headers=default_headers)
    assert_status_code(response, HTTP_BAD_REQUEST)
    assert_valid_schema(response.json(), 'error.json')

    developerAPI.delete_existing(test_developer['email'])


def test_developer_attributes():
    """
    Test create, read, update, delete attributes of developer
    """
    developerAPI = Developer(config)

    test_developer = developerAPI.create_new()

    developer_attributes_url = developerAPI.developer_url + '/' + test_developer['developerId'] + '/attributes'
    run_attributes_test(config, developer_attributes_url)

    developerAPI.delete_existing(test_developer['developerId'])


def test_developer_field_lengths():
    """
    Test field lengths
    """
    # TODO (erikbos)


# Negative tests

def test_get_all_developers_no_auth():
    """
    Test get all developers without basic auth
    """
    developerAPI = Developer(config)

    response = requests.get(developerAPI.developer_url)
    assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_get_developer_no_auth():
    """
    Test developer without basic auth
    """
    developerAPI = Developer(config)

    response = requests.get(developerAPI.developer_url + "/example@test.com")
    assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_get_all_developers_wrong_content_type():
    """
    Test get all developers, non-json content type
    """
    developerAPI = Developer(config)

    wrong_header = {'accept': 'application/unknown'}
    response = requests.get(developerAPI.developer_url, auth=developerAPI._get_auth(), headers=wrong_header)
    assert_status_code(response, HTTP_BAD_CONTENT)


def test_create_developer_wrong_content_type():
    """
    Test create developer, non-json content type
    """
    developerAPI = Developer(config)

    developer = {
        "email" : "example@test.com",
        "firstName" : "joe",
        "lastName" : "smith"
    }
    wrong_header = {'accept': 'application/unknown'}
    response = requests.post(developerAPI.developer_url, auth=developerAPI._get_auth(), headers=wrong_header, json=developer)
    assert_status_code(response, HTTP_BAD_CONTENT)


def test_create_developer_ignore_provided_fields():
    """
    Test create developer and try to set fields that cannot be set
    """
    developerAPI = Developer(config)

    r = random.randint(0,99999)
    email = developerAPI.generate_email_address(r)
    new_developer = {
        "email": email,
        "developerId": f"ID{r}",
        "firstName": f"first{r}",
        "lastName": f"last{r}",
        "userName": f"username{r}",
        "createdAt": 1562687288865,
        "createdBy": email,
        "lastModifiedAt": 1562687288865,
        "lastModifiedBy": email
    }
    created_developer = developerAPI.create_new(new_developer)

    # Not all provided fields must be accepted
    assert created_developer['developerId'] != new_developer['developerId']
    assert created_developer['createdAt'] != new_developer['createdAt']
    assert created_developer['createdBy'] != new_developer['createdBy']
    assert created_developer['lastModifiedAt'] != new_developer['lastModifiedAt']
    assert created_developer['lastModifiedBy'] != new_developer['lastModifiedBy']

    developerAPI.delete_existing(created_developer['developerId'])


def cleanup_test_developers():
    """
    Delete all leftover developers of partial executed tests
    """
    developerAPI = Developer(config)

    developerAPI.delete_all_test_developer()
