import os
import requests
import random
import json
# from common import assert_valid_schema, assert_status_code, assert_content_type_json
from common import *
from attribute import test_attributes

config = get_config()
base_url = config['api_url']
user = config['api_username']
password = config['api_password']

# Number of developers to create while CRUDing
developer_crud_test_count = 10

def test_get_all_developers_list():
     """
     Test get all developers, returns array of email addresses of developers
     """
     response = requests.get(base_url, auth=(user, password), headers=default_headers)
     assert_status_code(response, HTTP_OK)
     assert_content_type_json(response)
     assert_valid_schema(response.json(), 'developers-email-addresses.json')


def test_get_all_developers_list_detailed():
     """
     Test get all developers, returns array with zero or more developers
     """
     response = requests.get(base_url + '?expand=true', auth=(user, password), headers=default_headers)
     assert_status_code(response, HTTP_OK)
     assert_content_type_json(response)
     assert_valid_schema(response.json(), 'developers.json')
     # TODO (erikbos) testing of paginating


def test_crud_lifecycle_one_developer():
     """
     Test create, read, update, delete one developer
     """
     created_developer = create_new_developer()

     # Creating same, now existing developer, must not be possible
     create_existing_developer(created_developer)

     # Read existing developer by email
     retrieved_developer = get_existing_developer(created_developer['email'])
     assert_compare_developer(retrieved_developer, created_developer)

     # Read existing developer by developerid
     retrieved_developer = get_existing_developer(created_developer['developerId'])
     assert_compare_developer(retrieved_developer, created_developer)

     # Update email address while preserving developerID
     updated_developer = created_developer
     # updated_developer['email'] = 'newemailaddress@test.com'
     # updated_developer['attributes'] = [
     #           {
     #                "name" : "Status"
     #           },
     #           {
     #                "name" : "Shoesize",
     #                "value" : "42"
     #           }
     #      ]
     update_existing_developer(updated_developer)

     # Read updated developer by developerid
     retrieved_developer = get_existing_developer(updated_developer['email'])
     assert_compare_developer(retrieved_developer, updated_developer)

     # Delete just created developer
     deleted_developer = delete_existing_developer(updated_developer['email'])
     assert_compare_developer(deleted_developer, updated_developer)

     # Try to delete developer once more, must not exist anymore
     delete_nonexisting_developer(updated_developer['email'])


def test_crud_lifecycle_ten_developers():
     """
     Test create, read, update, delete ten developers
     """

     # Create developers
     created_developers = []
     for i in range(developer_crud_test_count):
          created_developers.append(create_new_developer())

     # Creating them once more must not be possible
     random.shuffle(created_developers)
     for i in range(developer_crud_test_count):
          create_existing_developer(created_developers[i])

     # Read created developers in random order
     random.shuffle(created_developers)
     for i in range(developer_crud_test_count):
          retrieved_developer = get_existing_developer(created_developers[i]['developerId'])
          assert_compare_developer(retrieved_developer, created_developers[i])

     # Delete each created developer
     random.shuffle(created_developers)
     for i in range(developer_crud_test_count):
          deleted_developer = delete_existing_developer(created_developers[i]['developerId'])
          # Check if just deleted developer matches with as created developer
          assert_compare_developer(deleted_developer, created_developers[i])

          # Try to delete developer once more, must not exist anymore
          delete_nonexisting_developer(created_developers[i]['developerId'])


def test_change_developer_status():
     """
     Test changing status of developer
     """
     test_developer = create_new_developer()

     # Change status to active
     url = base_url + '/' + test_developer['email'] + '?action=active'
     response = requests.post(url, auth=(user, password), headers=default_headers)
     assert_status_code(response, 204)
     assert response.content == b''

     assert get_existing_developer(test_developer['email'])['status'] == 'active'

     # Change status to inactive
     url = base_url + '/' + test_developer['email'] + '?action=inactive'
     response = requests.post(url, auth=(user, password), headers=default_headers)
     assert_status_code(response, 204)
     assert response.content == b''

     assert get_existing_developer(test_developer['email'])['status'] == 'inactive'

     # Change status to unknown value is unsupported
     url = base_url + '/' + test_developer['email'] + '?action=unknown'
     response = requests.post(url, auth=(user, password), headers=default_headers)
     assert_status_code(response, HTTP_BAD_REQUEST)
     assert_valid_schema(response.json(), 'error.json')

     delete_existing_developer(test_developer['email'])


def test_developer_attributes():
     """
     Test create, read, update, delete attributes of developer
     """
     test_developer = create_new_developer()

     developer_attributes_url = base_url + '/' + test_developer['developerId'] + '/attributes'
     test_attributes(config, developer_attributes_url)

     delete_existing_developer(test_developer['developerId'])


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
     response = requests.get(base_url)
     assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_get_developer_no_auth():
     """
     Test developer without basic auth
     """
     response = requests.get(base_url + "/example@test.com")
     assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_get_all_developers_wrong_content_type():
     """
     Test get all developers, non-json content type
     """
     wrong_header = {'accept': 'application/unknown'}
     response = requests.get(base_url, auth=(user, password), headers=wrong_header)
     assert_status_code(response, HTTP_BAD_CONTENT)


def test_create_developer_wrong_content_type():
     """
     Test create developer, non-json content type
     """
     developer = {
          "email" : "example@test.com",
          "firstName" : "joe",
          "lastName" : "smith"
     }
     wrong_header = {'accept': 'application/unknown'}
     response = requests.post(base_url, auth=(user, password), headers=wrong_header, json=developer)
     assert_status_code(response, HTTP_BAD_CONTENT)


def test_create_developer_ignore_provided_fields():
     """
     Test create developer and try to set fields that cannot be set
     """
     r = random.randint(0,99999)
     email = f"testuser{r}@test.com"
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
     created_developer = create_new_developer(new_developer)

     # Not all provided fields must be accepted
     assert created_developer['developerId'] != new_developer['developerId']
     assert created_developer['createdAt'] != new_developer['createdAt']
     assert created_developer['createdBy'] != new_developer['createdBy']
     assert created_developer['lastModifiedAt'] != new_developer['lastModifiedAt']
     assert created_developer['lastModifiedBy'] != new_developer['lastModifiedBy']

     delete_existing_developer(created_developer['developerId'])


def cleanup_test_developers():
     """
     Delete all leftover developers of partial executed tests
     """

     for i in range(developer_crud_test_count):
          new_developer = {
               "email" : f"testuser{i}@test.com",
          }
          developer_url = base_url + '/' + new_developer['email']
          requests.delete(developer_url, auth=(user, password))


def create_new_developer(new_developer=None):
     """
     Create new developer to be used as test subject. If none provided generate random developer data
     """

     if new_developer == None:
          r = random.randint(0,99999)
          new_developer = {
               "email" : f"testuser{r}@test.com",
               "firstName" : f"first{r}",
               "lastName" : f"last{r}",
               "userName" : f"username{r}",
               "attributes" : [ ],
          }

     response = requests.post(base_url, auth=(user, password), headers=default_headers, json=new_developer)
     assert_status_code(response, HTTP_CREATED)
     assert_content_type_json(response)

     # Check if just created developer matches with what we requested
     created_developer = response.json()
     assert_valid_schema(created_developer, 'developer.json')
     assert_compare_developer(created_developer, new_developer)

     return created_developer


def create_existing_developer(developer):
     """
     Attempt to create new developer which already exists
     """

     response = requests.post(base_url, auth=(user, password), headers=default_headers, json=developer)
     assert_status_code(response, HTTP_BAD_REQUEST)
     assert_content_type_json(response)
     assert_valid_schema(response.json(), 'error.json')


def update_existing_developer(developer):
     """
     Update existing developer
     """
     developer_url = base_url + '/' + developer['developerId']
     response = requests.post(developer_url, auth=(user, password), headers=default_headers, json=developer)
     assert_status_code(response, HTTP_OK)
     assert_content_type_json(response)

     updated_developer = response.json()
     assert_valid_schema(updated_developer, 'developer.json')

     return updated_developer

def get_existing_developer(id):
     """
     Get developer
     """
     developer_url = base_url + '/' + id
     response = requests.get(developer_url, auth=(user, password), headers=default_headers)
     assert_status_code(response, HTTP_OK)
     assert_content_type_json(response)
     retrieved_developer = response.json()
     assert_valid_schema(retrieved_developer, 'developer.json')

     return retrieved_developer


def delete_existing_developer(id):
     """
     Delete existing developer
     """
     developer_url = base_url + '/' + id
     response = requests.delete(developer_url, auth=(user, password), headers=default_headers)
     assert_status_code(response, HTTP_OK)
     assert_content_type_json(response)
     deleted_developer = response.json()
     assert_valid_schema(deleted_developer, 'developer.json')

     return deleted_developer


def delete_nonexisting_developer(id):
     """
     Delete non-existing developer
     """
     developer_url = base_url + '/' + id
     response = requests.delete(developer_url, auth=(user, password), headers=default_headers)
     assert_status_code(response, HTTP_NOT_FOUND)
     assert_content_type_json(response)


def assert_compare_developer(a, b):
     """
     Compares minimum required fields that can be set of two developers
     """
     assert a['email'] == b['email']
     assert a['firstName'] == b['firstName']
     assert a['lastName'] == b['lastName']
