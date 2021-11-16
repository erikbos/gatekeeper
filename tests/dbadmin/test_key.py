"""
Test suite to validate key endpoints operations
"""
import random
import secrets
from common import get_config, get_http_session
from developer import Developer
from developerapp import Application
from key import Key
from apiproduct import APIproduct
from attribute import run_attribute_tests

config = get_config()
session = get_http_session(config)


def test_key_crud():
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

    # create new key with assigned product
    updated_application = application_api.create_key_positive(
        created_application['name'], created_product['name'])

    # Updated application should now have two keys (first key gets created with app creation)
    assert len(updated_application['credentials']) == 2

    # attempt to delete product, not possible yet as it is assigned to key
    product_api.delete_negative_bad_request(created_product['name'])

    # remove apiproduct from just created key
    # as we don't know which key has the product assigned we
    # will have to iterate over all keys and assigned products

    keys = updated_application['credentials']
    for key_index in range(len(keys)):
        for product_index in range(len(keys[key_index]['apiProducts'])):
            if (keys[key_index]['apiProducts'][product_index]['apiproduct'] ==
                created_product['name']):
                key_api.delete_positive(keys[key_index]['consumerKey'])

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])
    # product must be deleted after application is deleted(!)
    # (eventhough it was already not assigned to any key anymore)
    product_api.delete_positive(created_product['name'])


def test_key_crud_multiple():
    """
    Test create, read, update, delete multiple key
    """


def test_key_import_various_lengths():
    """
    Test create, read, update, delete importing key of various length
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # try various key sizes
    created_keys = []
    for key_size in [2, 10, 100, 500, 1000, 2000]:
        key_in_hex_length = int(key_size / 2)
        new_key = {
            'consumerKey': secrets.token_hex(key_in_hex_length),
            'consumerSecret': secrets.token_hex(key_in_hex_length),
        }
        created_key = key_api.create_positive(new_key)
        retrieved_key = key_api.get_positive(created_key['consumerKey'])
        key_api.assert_compare(retrieved_key, created_key)

        # Attempt to creating same, already existing key, must not be possible
        key_api.create_negative(created_key)

        created_keys.append(created_key['consumerKey'])

    # Check each imported key is in listed in retrieved app
    retrieved_application = application_api.get_positive(created_application['name'])
    all_retrieved_keys = [retrieved_application['credentials'][i]['consumerKey'] for i in range(len(retrieved_application['credentials']))]
    if not all(consumer_key in all_retrieved_keys for consumer_key in created_keys):
        assert()

    # clean up
    for i in range(len(retrieved_application['credentials'])):
        key_api.delete_positive(retrieved_application['credentials'][i]['consumerKey'])

    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_key_after_add_application_one_key_present():
    """
    Test if creation of application also adds one key
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # Application should have 1 key created together with application
    assert (len(created_application['credentials']) == 1)

    # compare
    retrieved_key = key_api.get_positive(created_application['credentials'][0]['consumerKey'])
    assert (created_application['credentials'][0] == retrieved_key)

    # clean up
    key_api.delete_positive(created_application['credentials'][0]['consumerKey'])
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_key_import_incomplete_key():
    """
    Test importing incomplete key
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # try importing empty or incomplete keys, should have both consumerKey and consumerSecret
    key_api.create_negative({ 'consumerKey': '' })
    key_api.create_negative({ 'consumerSecret': '' })
    key_api.create_negative({ 'consumerKey': secrets.token_hex(10) })
    key_api.create_negative({ 'consumerSecret': secrets.token_hex(10) })

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_key_update_ignore_provided_fields():
    """
    Test update key and try to set fields that cannot be set
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()
    app_first_key = created_application['credentials'][0]['consumerKey']

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # Overwritting consumerKey is not allowed
    random_int = random.randint(0,99999)
    key_wrong_payload = {
        'consumerKey': f"secret{random_int}",
    }
    key_api.update_negative(app_first_key, key_wrong_payload)

    # try updating fields
    key_changed = {
        'attributes': [
              {
                   "name" : "Shoesize",
                   "value" : "42"
              }
        ],
        'issuedAt': 4242,
    }
    updated_key = key_api.update_positive(app_first_key, key_changed)

    # updatable fields
    assert updated_key['attributes'] == key_changed['attributes']

    # Not all provided fields must be accepted
    assert updated_key['issuedAt'] != key_changed['issuedAt']

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['developerId'])


def test_key_change_status():
    """
    Test changing status of apikey
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # After application creation status of first key should be approved
    assert created_application['credentials'][0]['status'] == 'approved'

    # Change status to revoke and back to approve
    consumer_key = created_application['credentials'][0]['consumerKey']
    key_api.change_status_revoke_positive(consumer_key)
    assert key_api.get_positive(consumer_key)['status'] == 'revoked'

    key_api.change_status_approve_positive(consumer_key)
    assert key_api.get_positive(consumer_key)['status'] == 'approved'

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_key_api_product_add_remove_multiple():
    """
    Test add and remove multiple products to key
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()
    consumer_key = created_application['credentials'][0]['consumerKey']

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    product_api = APIproduct(config, session)
    created_product_1 = product_api.create_positive()

    # Add apiproduct to key and replace attributes of this key
    new_api_product_assigment1 = {
        "apiProducts": [
            created_product_1['name']
        ],
        "attributes": [
            {
            "name": "Location",
            "value": "NYC"
            }
        ]
    }
    key_api.update_positive(consumer_key, new_api_product_assigment1)
    retrieved_key = key_api.get_positive(consumer_key)
    assert (retrieved_key['apiProducts'][0]['apiproduct'] ==
                new_api_product_assigment1['apiProducts'][0])
    assert retrieved_key['attributes'] == new_api_product_assigment1['attributes']

    # Add a second apiproduct and replace attributes of this key
    created_product_2 = product_api.create_positive()
    new_api_product_assigment2 = {
        "apiProducts": [
            created_product_2['name']
        ],
        "attributes": [
            {
            "name": "Location",
            "value": "AMS"
            }
        ]
    }
    key_api.update_positive(consumer_key, new_api_product_assigment2)
    retrieved_key = key_api.get_positive(consumer_key)
    assert len(retrieved_key['apiProducts']) == 2
    assert retrieved_key['attributes'] == new_api_product_assigment2['attributes']

    # remove first apiproduct from key, second product should be there
    key_api.api_product_delete_positive(consumer_key, created_product_1['name'])
    retrieved_key = key_api.get_positive(consumer_key)
    assert (retrieved_key['apiProducts'][0]['apiproduct'] ==
                new_api_product_assigment2['apiProducts'][0])
    assert retrieved_key['attributes'] == new_api_product_assigment2['attributes']

    # remove second product
    key_api.api_product_delete_positive(consumer_key, created_product_2['name'])
    assert len(key_api.get_positive(consumer_key)['apiProducts']) == 0

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])
    product_api.delete_positive(created_product_1['name'])
    product_api.delete_positive(created_product_2['name'])


def test_key_api_product_change_status():
    """
    Test changing status of product assigned to apikey
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    # remove default provided key
    key_api.delete_positive(created_application['credentials'][0]['consumerKey'])

    # create product as we want to create a key with assigned product
    product_api = APIproduct(config, session)
    created_product = product_api.create_positive()

    # create new key with one assigned product
    updated_application = application_api.create_key_positive(
        created_application['name'], created_product['name'])
    consumer_key = updated_application['credentials'][0]['consumerKey']
    api_product_name = updated_application['credentials'][0]['apiProducts'][0]['apiproduct']

    # Change status to revoke(d) and back to approve(d)
    key_api.api_product_change_status(consumer_key, api_product_name, 'revoke', True)
    assert key_api.get_positive(consumer_key)['apiProducts'][0]['status'] == 'revoked'

    key_api.api_product_change_status(consumer_key, api_product_name, 'approve', True)
    assert key_api.get_positive(consumer_key)['apiProducts'][0]['status'] == 'approved'

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


def test_key_attributes():
    """
    Test create, read, update, delete attributes of key
    """
    developer_api = Developer(config, session)
    created_developer = developer_api.create_positive()

    application_api = Application(config, session, created_developer['email'])
    created_application = application_api.create_new()

    key_api = Key(config, session, created_developer['email'], created_application['name'])

    consumer_key = created_application['credentials'][0]['consumerKey']

    application_attributes_url = (application_api.application_url
                                  + '/' + created_application['name']
                                  + '/keys/' + consumer_key
                                  + '/attributes')

    run_attribute_tests(config, session, application_attributes_url)

    # clean up
    application_api.delete_positive(created_application['name'])
    developer_api.delete_positive(created_developer['email'])


# test keyexpiresin from developerapp

# test import same key twice

# test key for product with approvalType manual

# test posting apiproducts against existing key


# test posting full key details
# test setting all fields
# POST /v1/organizations/{organization}/developers/{developer_email_or_id}
#           /apps/{app_name}/keys/{consumer_key_name}
# {
#  "apiProducts": ["product_1", "product_2"]
# }
# test posting and put key to app endpoin
# apiproduct assignment
# test importing partial key
