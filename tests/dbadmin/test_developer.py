"""
Test suite to validate developer endpoints operations
"""
import copy
import random
from common import get_config, get_http_session, API, assert_status_code
from httpstatus import  HTTP_AUTHORIZATION_REQUIRED, HTTP_BAD_CONTENT
from developer import Developer
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_developer_get_all():
    """
    Test get all developers, returns array of email addresses of developers
    """
    developer_api = Developer(config, session)
    developer_api.get_all()


def test_developer_get_all_detailed():
    """
    Test get all developers, returns array with zero or more developers
    """
    developer_api = Developer(config, session)
    developer_api.get_all_detailed()


def test_developer_crud_one():
    """
    Test create, read, update, delete one developer
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    # Creating same, now existing developer, must not be possible
    developer_api.create_negative(created_developer)

    # Read existing developer by email
    retrieved_developer = developer_api.get_positive(created_developer['email'])
    developer_api.assert_compare(retrieved_developer, created_developer)

    # Read existing developer by developerid
    retrieved_developer = developer_api.get_positive(created_developer['developerId'])
    developer_api.assert_compare(retrieved_developer, created_developer)

    # Check if created developer shows up in global developer list
    if created_developer['email'] not in developer_api.get_all():
        assert()

    # Update to new email address and attributes
    updated_developer = copy.deepcopy(created_developer)
    random_int = random.randint(0,99999)
    updated_developer['email'] = f'newemailaddress{random_int}@test.com'
    updated_developer['attributes'] = [
              {
                   "name" : "Status",
                   "value" : ""
              },
              {
                   "name" : "Shoesize",
                   "value" : "42"
              }
         ]
    developer_api.update_positive(created_developer['email'], updated_developer)

    # Read updated developer by developerid
    retrieved_developer = developer_api.get_positive(updated_developer['developerId'])
    developer_api.assert_compare(retrieved_developer, updated_developer)

    # Check if updated developer shows up in global developer list
    if updated_developer['email'] not in developer_api.get_all():
        assert()

    # Delete just created developer
    deleted_developer = developer_api.delete_positive(updated_developer['email'])
    developer_api.assert_compare(deleted_developer, updated_developer)

    # Try to delete developer once more, must not exist anymore
    developer_api.delete_negative_notfound(updated_developer['email'])


def test_developer_crud_multiple():
    """
    Test create, read, update, delete multiple developers
    """
    developer_api = Developer(config, session)

    # Create developers
    created_developers = []
    for i in range(config['entity_count']):
        created_developers.append(developer_api.create_positive())

    # Creating them once more must not be possible
    random.shuffle(created_developers)
    for i in range(config['entity_count']):
        developer_api.create_negative(created_developers[i])

    # Read created developers in random order
    random.shuffle(created_developers)
    for i in range(config['entity_count']):
        retrieved_developer = developer_api.get_positive(created_developers[i]['developerId'])
        developer_api.assert_compare(retrieved_developer, created_developers[i])

    # Check if all created developers shows up in global developers list
    all_created = [created_developers[i]['email'] for i in range(config['entity_count'])]
    all_developers = developer_api.get_all()
    if not all(email_address in all_developers for email_address in all_created):
        assert()

    # Delete each created developer
    random.shuffle(created_developers)
    for i in range(config['entity_count']):
        deleted_developer = developer_api.delete_positive(created_developers[i]['developerId'])
        # Check if just deleted developer matches with as created developer
        developer_api.assert_compare(deleted_developer, created_developers[i])

        # Try to delete developer once more, must not exist anymore
        developer_api.delete_negative_notfound(created_developers[i]['developerId'])


def test_developer_attributes():
    """
    Test create, read, update, delete attributes of developer
    """
    developer_api = Developer(config, session)

    test_developer = developer_api.create_positive()

    developer_attributes_url = (developer_api.developer_url +
        '/' + test_developer['email'] + '/attributes')
    run_attribute_tests(config, session, developer_attributes_url)

    # clean up
    developer_api.delete_positive(test_developer['email'])


def test_developer_change_status():
    """
    Test changing status of developer
    """
    developer_api = Developer(config, session)
    test_developer = developer_api.create_positive()

    # Developer status should be active after creation
    assert test_developer['status'] == 'active'

    # Change status to inactive & active
    email = test_developer['email']
    developer_api.change_status_inactive_positive(email)
    assert developer_api.get_positive(email)['status'] == 'inactive'

    developer_api.change_status_active_positive(email)
    assert developer_api.get_positive(email)['status'] == 'active'

    developer_api.delete_positive(email)


def test_developer_field_lengths():
    """
    Test field lengths
    """
    # We want to test that we cannot send zero-length, too long fields etc
    # TODO test wrong field lengths


# Negative tests
def test_developer_get_all_no_auth():
    """
    Test get all developers without basic auth
    """
    developer_api = Developer(config, session)

    response = session.get(developer_api.developer_url, auth=())
    assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


def test_developer_get_no_auth():
    """
    Test developer without basic auth
    """
    developer_api = Developer(config, session)

    response = session.get(developer_api.developer_url + "/example@test.com", auth=())
    assert_status_code(response, HTTP_AUTHORIZATION_REQUIRED)


# def test_developer_get_all_wrong_content_type():
#     """
#     Test get all developers, non-json content type
#     """
#     developer_api = Developer(config, session)

#     wrong_header = {'accept': 'application/unknown'}
#     response = session.get(developer_api.developer_url, headers=wrong_header)
#     assert_status_code(response, HTTP_BAD_CONTENT)


# def test_developer_create_wrong_content_type():
#     """
#     Test create developer, non-json content type
#     """
#     developer_api = Developer(config, session)

#     developer = {
#         "email" : "example@test.com",
#         "firstName" : "joe",
#         "lastName" : "smith"
#     }
#     wrong_header = {'accept': 'application/unknown'}
#     response = session.post(developer_api.developer_url, headers=wrong_header, json=developer)
#     assert_status_code(response, HTTP_BAD_CONTENT)


def test_developer_create_ignore_provided_fields():
    """
    Test create developer and try to set fields that cannot be set
    """
    developer_api = Developer(config, session)

    random_int = random.randint(0,99999)
    email = developer_api.generate_email_address(random_int)
    new_developer = {
        "email": email,
        "developerId": f"ID{random_int}",
        "firstName": f"first{random_int}",
        "lastName": f"last{random_int}",
        "userName": f"username{random_int}",
        "attributes": [],
        "createdAt": 1562687288865,
        "createdBy": email,
        "lastModifiedAt": 1562687288865,
        "lastModifiedBy": email
    }
    created_developer = developer_api.create_positive(new_developer)

    print("Q ", created_developer)

    # Not all provided fields must be accepted
    assert created_developer['developerId'] != new_developer['developerId']
    assert created_developer['createdAt'] != new_developer['createdAt']
    assert created_developer['createdBy'] != new_developer['createdBy']
    assert created_developer['lastModifiedAt'] != new_developer['lastModifiedAt']
    assert created_developer['lastModifiedBy'] != new_developer['lastModifiedBy']

    developer_id = created_developer['developerId']

    # attempt to manipulate developer_id of existing developer to wrong value
    created_developer['developerId'] = developer_id + 'wrong'
    developer_api.update_positive(created_developer['email'], created_developer)

    # Retrieve developer, id should still be the original one
    change_developer = developer_api.get_positive(created_developer['email'])
    assert change_developer['developerId'] != created_developer['developerId']

    # clean up
    developer_api.delete_positive(developer_id)


def cleanup_test_developers():
    """
    Delete all leftover developers created by tests
    """
    developer_api = Developer(config, session)

    developer_api.delete_all_test_developer()
