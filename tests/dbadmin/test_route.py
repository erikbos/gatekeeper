"""
Test suite to validate routes endpoints operations
"""
import copy
import random
import urllib
from common import get_config, API
from route import Route
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_route_get_all():
    """
    Test get all routes
    """
    route_api = Route(config, session)
    route_api.get_all()


def test_route_crud_lifecycle():
    """
    Test create, read, update, delete one route
    """
    route_api = Route(config, session)
    created_route = route_api.create_positive()

    # Creating same, now existing route, must not be possible
    route_api.create_negative(created_route)

    # Read existing route by name
    retrieved_route = route_api.get_positive(created_route['name'])
    route_api.assert_compare(retrieved_route, created_route)

    # Check if created route shows up in global route list
    assert(True for l in route_api.get_all()['route'] if l['name'] == created_route['name'])

    # Update to new email address and attributes
    updated_route = copy.deepcopy(created_route)
    random_int = random.randint(0,99999)
    updated_route['displayName'] = f'Testsuite Route Name {random_int}'
    updated_route['attributes'] = [
        {
            "name": "HostHeader",
            "value": "42.example.com"
        }
    ]
    route_api.update_positive(created_route['name'], updated_route)

    # Check if updated route shows up in global route list
    assert(True for l in route_api.get_all()['route'] if l['name'] == created_route['name'])

    # Delete just created route
    deleted_route = route_api.delete_positive(updated_route['name'])
    route_api.assert_compare(deleted_route, updated_route)

    # Try to delete route once more, must not exist anymore
    route_api.delete_negative(updated_route['name'])
