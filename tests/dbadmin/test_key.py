"""
Test suite to validate key endpoints operations
"""
import secrets
from common import get_config, get_http_session
from developer import Developer
from application import Application
from key import Key
from attribute import run_attribute_tests

config = get_config()
session = get_http_session(config)


def test_key_crud_lifecycle():
    """
    Test create, read, update, delete keys of one application
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_new()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # Validate retrieve first key that is auto-generated when app is created
    app_first_key = created_application['credentials'][0]
    print(app_first_key)
    first_key = key_api.get_existing(app_first_key['consumerKey'])
    key_api.assert_compare(first_key, app_first_key)

    # create a new key
    new_key = {
        'consumerKey': secrets.token_hex(16),
        'consumerSecret': secrets.token_hex(16),
    }
    created_key = key_api.create(new_key, True)
    retrieved_key = key_api.get_existing(created_key['consumerKey'])
    key_api.assert_compare(retrieved_key, created_key)

    # Attempt to creating same, already existing application, must not be possible
    key_api.create(created_key, False)

    # Validate just added key is part of app response set
    retrieved_application = application_api.get_existing(created_application['name'])
    key_present = False
    for i in range(len(retrieved_application['credentials'])):
        if retrieved_application['credentials'][i]['consumerKey'] == new_key['consumerKey']:
            key_present = True
    assert key_present

    # clean up
    application_api.delete_existing(created_application['name'])
    developer_api.delete_existing(created_developer['email'])

# test crud life cycle multiple

def test_key_attributes():
    """
    Test create, read, update, delete attributes of key
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_new()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    # Each new app has one auto-generated key, test against that key's attribute endpoint
    key_attributes_url = (config['api_url'] +
                            '/developers/' + created_developer['email'] +
                            '/apps/' + created_application['name'] +
                            '/keys/' + created_application['credentials'][0]['consumerKey'] +
                            '/attributes')
    run_attribute_tests(config, session, key_attributes_url)

    # clean up
    application_api.delete_existing(created_application['name'])
    developer_api.delete_existing(created_developer['email'])

# test various key sizes

# test setting all fields

# test deleting app while it has a key

# test key import /organizations/{org_name}/developers/{developer_email}/apps/{app_name}/keys/create
# POST /v1/organizations/{organization}/developers/{developer_email_or_id}/apps/{app_name}/keys/create
# {
#   "consumerKey": "key",
#   "consumerSecret": "secret"
# }

# test auto key creation after new app creation

# test change key status
# POST /organizations/{org_name}/developers/{developer_email}/apps/{app_name}/keys/{consumer_key}
# action=approve / revoke

# apiproduct assignment

# test adding product to key
# POST /v1/organizations/{organization}/developers/{developer_email_or_id}/apps/{app_name}/keys/{consumer_key_name}
# {
#  "apiProducts": ["product_1", "product_2"]
# }

# test removing apiproduct

# test change apiproduct status per key
# POST /organizations/{org_name}/developers/{developer_email}/apps/
#           {app_name}/keys/{consumer_key}/apiproducts/{apiproduct_name}
# action=approve / revoke
