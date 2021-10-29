"""
Test suite to validate key endpoints operations
"""
import secrets
from common import get_config, get_http_session
from developer import Developer
from developerapp import Application
from key import Key
from apiproduct import APIproduct

config = get_config()
session = get_http_session(config)


def test_key_crud_lifecycle():
    """
    Test create, read, update, delete new key to application
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # create product as we cannot create a key without assigned product
    product_api = APIproduct(config, session)
    created_product = product_api.create_positive()

    # print("created_product -> ",  created_product)

    # new key can be assigned via application (really !)
    updated_application = application_api.create_key_positive(
        created_application['name'], created_product['name'])

    # Updated application should now have two keys (first one gets created with app creation)
    assert len(updated_application['credentials']) == 2

    # attempt to delete product, not possible yet as it is assigned to key
    product_api.delete_negative_bad_request(created_product['name'])

    # remove apiproduct from just created key
    # as we don't know which key it has assign we will have to iterate over all keys and assigned products
    for key_index in range(len(updated_application['credentials'])):
        for product_index in range(len(updated_application['credentials'][key_index]['apiProducts'])):
            if updated_application['credentials'][key_index] \
                ['apiProducts'][product_index]['apiproduct'] == created_product['name']:
                key_api.delete_positive(updated_application['credentials'][key_index]['consumerKey'])

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])
    # product must be deleted after application delete
    # eventhough it was not assigned to any key anymore
    product_api.delete_positive(created_product['name'])


def test_key_import_various_lengths():
    """
    Test create, read, update, delete importing key of various length
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # Validate retrieve first key that is auto-generated when app is created
    app_first_key = created_application['credentials'][0]
    first_key = key_api.get_positive(app_first_key['consumerKey'])
    key_api.assert_compare(first_key, app_first_key)

    for key_size in [1, 10, 100, 500, 1000, 2000]:
        # create a new key
        new_key = {
            'consumerKey': secrets.token_hex(key_size),
            'consumerSecret': secrets.token_hex(key_size),
        }
        created_key = key_api.create(new_key, True)
        retrieved_key = key_api.get_positive(created_key['consumerKey'])
        key_api.assert_compare(retrieved_key, created_key)

        # Attempt to creating same, already existing key, must not be possible
        key_api.create(created_key, False)

        # Validate just imported key is part of app response set
        retrieved_application = application_api.get_positive(created_application['name'])
        key_present = False
        for i in range(len(retrieved_application['credentials'])):
            if retrieved_application['credentials'][i]['consumerKey'] == new_key['consumerKey']:
                key_present = True
        assert key_present


    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_change_key_status():
    """
    Test changing status of apikey
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # After app creation status of first key is approved
    assert created_application['credentials'][0]['status'] == 'approved'

    # Change status to revoke and back to approve
    first_key = created_application['credentials'][0]['consumerKey']
    key_api.change_status(first_key, 'revoke', True)
    assert key_api.get_positive(first_key)['status'] == 'revoked'

    key_api.change_status(first_key, 'approve', True)
    assert key_api.get_positive(first_key)['status'] == 'approved'

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])



# apiproduct assignment

# test adding product to key
# POST /v1/organizations/{organization}/developers/{developer_email_or_id}
#           /apps/{app_name}/keys/{consumer_key_name}
# {
#  "apiProducts": ["product_1", "product_2"]
# }

# test removing apiproduct

# test change apiproduct status per key
# POST /organizations/{org_name}/developers/{developer_email}/apps/
#           {app_name}/keys/{consumer_key}/apiproducts/{apiproduct_name}
# action=approve / revoke

# test setting all fields

# test deleting app while it has a key

# test deleting app while it has a key
