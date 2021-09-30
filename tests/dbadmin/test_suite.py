import requests
import random
from common import assert_valid_schema_error, assert_status_code, get_config
from httpstatus import HTTP_NO_CONTENT, HTTP_AUTHORIZATION_REQUIRED, HTTP_BAD_REQUEST, HTTP_BAD_CONTENT
from developer import Developer
from attribute import run_attribute_tests

config = get_config()

session = requests.Session()
session.auth = (config['api_username'], config['api_password'])
session.headers = {
    'accept': 'application/json',
      'user-agent': 'Gatekeeper testsuite'
    }


def test_developer_get_all():
    """
    Test get all developers, returns array of email addresses of developers
    """
    developerAPI = Developer(config, session)
    developerAPI.get_all()


def test_developer_list_detailed():
    """
    Test get all developers, returns array with zero or more developers
    """
    developerAPI = Developer(config, session)
    developerAPI.get_all_detailed()


def test_developer_crud_lifecycle():
    """
    Test create, read, update, delete one developer
    """
    developerAPI = Developer(config, session)

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


def test_developer_crud_lifecycle_ten_random():
    """
    Test create, read, update, delete ten randomly named developers
    """
    developerAPI = Developer(config, session)

    # Create developers
    created_developers = []
    for i in range(config['entity_count']):
        created_developers.append(developerAPI.create_new())

    # Creating them once more must not be possible
    random.shuffle(created_developers)
    for i in range(config['entity_count']):
        developerAPI.create_existing_developer(created_developers[i])

    # Read created developers in random order
    random.shuffle(created_developers)
    for i in range(config['entity_count']):
        retrieved_developer = developerAPI.get_existing(created_developers[i]['developerId'])
        developerAPI.assert_compare_developer(retrieved_developer, created_developers[i])

    # Delete each created developer
    random.shuffle(created_developers)
    for i in range(config['entity_count']):
        deleted_developer = developerAPI.delete_existing(created_developers[i]['developerId'])
        # Check if just deleted developer matches with as created developer
        developerAPI.assert_compare_developer(deleted_developer, created_developers[i])

        # Try to delete developer once more, must not exist anymore
        developerAPI.delete_nonexisting(created_developers[i]['developerId'])


def test_developer_change_status():
    """
    Test changing status of developer
    """
    developerAPI = Developer(config, session)

    test_developer = developerAPI.create_new()

    # Change status to active
    url = developerAPI.developer_url + '/' + test_developer['email'] + '?action=active'
    response = session.post(url)
    assert_status_code(response, HTTP_NO_CONTENT)
    assert response.content == b''

    assert developerAPI.get_existing(test_developer['email'])['status'] == 'active'

    # Change status to inactive
    url = developerAPI.developer_url + '/' + test_developer['email'] + '?action=inactive'
    response = session.post(url)
    assert_status_code(response, HTTP_NO_CONTENT)
    assert response.content == b''

    assert developerAPI.get_existing(test_developer['email'])['status'] == 'inactive'

    # Change status to unknown value is unsupported
    url = developerAPI.developer_url + '/' + test_developer['email'] + '?action=unknown'
    response = session.post(url)
    assert_status_code(response, HTTP_BAD_REQUEST)
    assert_valid_schema_error(response.json())

    developerAPI.delete_existing(test_developer['email'])


def test_developer_attributes():
    """
    Test create, read, update, delete attributes of developer
    """
    developerAPI = Developer(config, session)

    test_developer = developerAPI.create_new()

    developer_attributes_url = developerAPI.developer_url + '/' + test_developer['developerId'] + '/attributes'
    run_attribute_tests(config, session, developer_attributes_url)

    developerAPI.delete_existing(test_developer['developerId'])


def test_developer_field_lengths():
    """
    Test field lengths
    """
    # We want to test that we cannot send zero-length, too long fields etc
    # TODO (erikbos)


# Negative tests
def test_developer_get_all_no_auth():
    """
    Test get all developers without basic auth
    """
    developerAPI = Developer(config, session)

    response = session.get(developerAPI.developer_url, auth=())
    assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_developer_get_no_auth():
    """
    Test developer without basic auth
    """
    developerAPI = Developer(config, session)

    response = session.get(developerAPI.developer_url + "/example@test.com", auth=())
    assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_developers_get_all_wrong_content_type():
    """
    Test get all developers, non-json content type
    """
    developerAPI = Developer(config, session)

    wrong_header = {'accept': 'application/unknown'}
    response = session.get(developerAPI.developer_url, headers=wrong_header)
    assert_status_code(response, HTTP_BAD_CONTENT)


def test_developer_create_wrong_content_type():
    """
    Test create developer, non-json content type
    """
    developerAPI = Developer(config, session)

    developer = {
        "email" : "example@test.com",
        "firstName" : "joe",
        "lastName" : "smith"
    }
    wrong_header = {'accept': 'application/unknown'}
    response = session.post(developerAPI.developer_url, headers=wrong_header, json=developer)
    assert_status_code(response, HTTP_BAD_CONTENT)


def test_developer_create_ignore_provided_fields():
    """
    Test create developer and try to set fields that cannot be set
    """
    developerAPI = Developer(config, session)

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
    developerAPI = Developer(config, session)

    developerAPI.delete_all_test_developer()
