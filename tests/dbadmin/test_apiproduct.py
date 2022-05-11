"""
Test suite to validate api product endpoints operations
"""
import copy
import random
from common import get_config, get_http_session, API
from apiproduct import APIproduct
from attribute import run_attribute_tests

config = get_config()
session = API(config, '../../openapi/gatekeeper.yaml')


def test_api_product_get_all():
    """
    Test get all developers, returns array of email addresses of developers
    """
    product_api = APIproduct(config, session)
    product_api.get_all()


def test_api_product_get_all_detailed():
    """
    Test get all developers, returns array with zero or more developers
    """
    product_api = APIproduct(config, session)
    product_api.get_all_detailed()


def test_api_product_crud_lifecycle():
    """
    Test create, read, update, delete one api product
    """
    product_api = APIproduct(config, session)
    created_product = product_api.create_positive()

    # Creating same, now existing developer, must not be possible
    product_api.create_negative(created_product)

    # Read existing developer by name
    retrieved_product = product_api.get_positive(created_product['name'])
    product_api.assert_compare(retrieved_product, created_product)

    # Check if created product shows up in global apiproduct list
    if created_product['name'] not in product_api.get_all():
        assert()

    # Update to new email address and attributes
    updated_product = copy.deepcopy(created_product)
    random_int = random.randint(0,99999)
    updated_product['displayName'] = f'Testsuite Product Name {random_int}'
    updated_product['attributes'] = [
              {
                   "name" : "Test",
              },
              {
                   "name" : "Status",
                   "value" : "Royal"
              }
         ]
    product_api.update_positive(created_product['name'], updated_product)

    # Check if updated product shows up in global product list
    if updated_product['name'] not in product_api.get_all():
        assert()

    # Delete just created product
    deleted_product = product_api.delete_positive(updated_product['name'])
    product_api.assert_compare(deleted_product, updated_product)

    # Try to delete product once more, must not exist anymore
    product_api.delete_negative(updated_product['name'])


def test_api_product_attributes():
    """
    Test create, read, update, delete attributes of api product
    """
    product_api = APIproduct(config, session)

    test_api_product = product_api.create_positive()

    product_attributes_url = (product_api.api_product_url +
        '/' + test_api_product['name'] + '/attributes')
    run_attribute_tests(config, session, product_attributes_url)

    # clean up
    product_api.delete_positive(test_api_product['name'])
