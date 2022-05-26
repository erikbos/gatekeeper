"""
Test suite to validate listeners endpoints operations
"""
import copy
import random
import urllib
from common import get_config, API
from listener import Listener
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_listener_get_all():
    """
    Test get all listeners
    """
    listener_api = Listener(config, session)
    listener_api.get_all()


def test_listener_crud_lifecycle():
    """
    Test create, read, update, delete one listener
    """
    listener_api = Listener(config, session)
    created_listener = listener_api.create_positive()

    # Creating same, now existing listener, must not be possible
    listener_api.create_negative(created_listener)

    # Read existing listener by name
    retrieved_listener = listener_api.get_positive(created_listener['name'])
    listener_api.assert_compare(retrieved_listener, created_listener)

    # Check if created listener shows up in global listener list
    assert(True for l in listener_api.get_all()['listener'] if l['name'] == created_listener['name'])

    # Update to new email address and attributes
    updated_listener = copy.deepcopy(created_listener)
    random_int = random.randint(0,99999)
    updated_listener['displayName'] = f'Testsuite Listener Name {random_int}'
    updated_listener['attributes'] = [
        {
            "name": "IdleTimeout",
            "value": "18ms"
        },
        {
            "name": "ServerName",
            "value": "eb"
        }
    ]
    listener_api.update_positive(created_listener['name'], updated_listener)

    # Check if updated listener shows up in global listener list
    assert(True for l in listener_api.get_all()['listener'] if l['name'] == created_listener['name'])

    # Delete just created listener
    deleted_listener = listener_api.delete_positive(updated_listener['name'])
    listener_api.assert_compare(deleted_listener, updated_listener)

    # Try to delete listener once more, must not exist anymore
    listener_api.delete_negative(updated_listener['name'])
