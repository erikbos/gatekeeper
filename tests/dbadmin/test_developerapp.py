"""
Test suite to validate developer application endpoints operations
"""
import copy
import random
from common import get_config, API
from developer import Developer
from developerapp import Application
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_application_get_all():
    """
    Test get all applications
    """
    application_api = Application(config, session, None)
    application_api.get_all_global()


def test_application_get_all_detailed():
    """
    Test get all applications detailed
    """
    application_api = Application(config, session, None)
    application_api.get_all_global()


def test_application_crud_one():
    """
    Test create, read, update, delete one application of one developer
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    # Attempt to creating same, already existing application, must not be possible
    application_api.create_negative(created_application)

    # Read existing app by name
    retrieved_application = application_api.get_positive(created_application['name'])
    application_api.assert_compare(retrieved_application, created_application)

    # Update name and attributes
    updated_application = copy.deepcopy(created_application)
    updated_application['displayName'] = 'Test change of app name'
    updated_application['attributes'] = [
              {
                   "name" : "Test",
                   "value" : ""
              },
              {
                   "name" : "Status",
                   "value" : "Royal"
              }
         ]
    updated_application['callbackUrl'] = ''
    application_api.update_positive(updated_application)

    # Read updated application by name
    retrieved_application = application_api.get_positive(updated_application['name'])
    application_api.assert_compare(retrieved_application, updated_application)

    # Read updated application by uuid on global app endpoint
    retrieved_application_uuid = application_api.get_by_uuid_positive(updated_application['appId'])
    application_api.assert_compare(retrieved_application_uuid, updated_application)

    # Check if application name shows up inside developer's object
    retrieved_developer = developer_api.get_positive(created_developer['email'])
    if created_application['name'] not in retrieved_developer['apps']:
        assert()

    # Check if application shows up in global application list
    if created_application['appId'] not in application_api.get_all_global():
        assert()

    # Check if application shows up in developer's application list
    if created_application['name'] not in application_api.get_all():
        assert()

    # Delete just created application
    deleted_application = application_api.delete_positive(updated_application['name'])
    application_api.assert_compare(deleted_application, updated_application)

    # Try to delete application once more, must not exist anymore
    application_api.delete_negative(updated_application['name'])

    # clean up
    developer_api.delete_positive(created_developer['email'])


def test_application_crud_multiple():
    """
    Test create, read, update, delete multiple applications, randomly named
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])

    # Create applications
    created_applications = []
    for i in range(config['entity_count']):
        created_applications.append(application_api.create_new())

    # Creating them once more must not be possible
    random.shuffle(created_applications)
    for i in range(config['entity_count']):
        application_api.create_negative(created_applications[i])

    # Read created application in random order
    random.shuffle(created_applications)
    for i in range(config['entity_count']):
        retrieved_application = application_api.get_positive(created_applications[i]['name'])
        application_api.assert_compare(retrieved_application, created_applications[i])

    # Check each app is listed in global app list
    all_created = [created_applications[i]['appId'] for i in range(config['entity_count'])]
    all_apps = application_api.get_all_global()
    if not all(app_uuid in all_apps for app_uuid in all_created):
        assert()

    # Delete each created application
    random.shuffle(created_applications)
    for i in range(config['entity_count']):
        deleted_application = application_api.delete_positive(created_applications[i]['name'])
        # Check if just deleted application matches with as created application
        application_api.assert_compare(deleted_application, created_applications[i])

        # Try to delete developer once more, must not exist anymore
        application_api.delete_negative(created_applications[i]['name'])

    # clean up
    developer_api.delete_positive(created_developer['email'])


def test_application_attributes():
    """
    Test create, read, update, delete attributes of application
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    application_attributes_url = (application_api.url +
                                '/' + created_application['name'] + '/attributes')
    run_attribute_tests(config, session, application_attributes_url)

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_application_change_status():
    """
    Test changing status of application
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    # application status should be approved after creation
    assert created_application['status'] == 'approved'

    # try revoking followed by approving
    app_name = created_application['name']
    application_api.change_status_revoke_positive(app_name)
    assert application_api.get_positive(app_name)['status'] == 'revoked'

    application_api.change_status_approve_positive(app_name)
    assert application_api.get_positive(app_name)['status'] == 'approved'

    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_developer_attempt_delete_with_one_application():
    """
    Attempt to delete developer with one application, should fail
    """

    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    # we expect delete developer to fail
    developer_api.delete_negative_badrequest(created_developer['email'])

    # cleanup
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_application_field_lengths():
    """
    Test field lengths
    """
    # We want to test that we cannot send zero-length, too long fields etc
    # TODO test wrong field lengths

## TODO two developer each "My First App"
