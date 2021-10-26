"""
Apiproduct module does all REST API operations on api_product endpoint
"""
import random
from common import assert_valid_schema, assert_status_code, \
    assert_content_type_json, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Apiproduct:
    """
    Apiproduct does all REST API operations on apiproduct endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.api_product_url = config['api_url'] + '/apiproducts'
        self.schemas = {
            'api_product': load_json_schema('api_product.json'),
            'error': load_json_schema('error.json'),
        }


    def generate_api_product_name(self, number):
        """
        Returns name of test api product based upon provided number
        """
        return f"testsuite-apiproduct{number}"


    def create_new(self, new_api_product=None):
        """
        Create new api product to be used as test subject. If none provided generate random API product
        """

        if new_api_product is None:
            random_int = random.randint(0,99999)
            new_api_product = {
                "name" : self.generate_api_product_name(random_int),
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.api_product_url, json=new_api_product)
        assert_status_code(response, HTTP_CREATED)
        assert_content_type_json(response)

        # Check if just created api product matches with what we requested
        created_api_product = response.json()
        assert_valid_schema(created_api_product, self.schemas['api_product'])
        self.assert_compare(created_api_product, new_api_product)

        return created_api_product


    def create_existing(self, api_product):
        """
        Attempt to create api product which already exists, should fail
        """
        response = self.session.post(self.api_product_url, json=api_product)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['error'])


    def get_existing(self, api_product):
        """
        Get existing api product
        """
        api_product_url = self.api_product_url + '/' + api_product
        response = self.session.get(api_product_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_api_product = response.json()
        assert_valid_schema(retrieved_api_product, self.schemas['api_product'])

        return retrieved_api_product


    def update_existing(self, api_product, updated_api_product):
        """
        Update existing api product
        """
        api_product_url = self.api_product_url + '/' + api_product
        response = self.session.post(api_product_url, json=updated_api_product)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_api_product = response.json()
        assert_valid_schema(updated_api_product, self.schemas['api_product'])

        return updated_api_product


    def delete_existing(self, api_product):
        """
        Delete existing api product
        """
        api_product_url = self.api_product_url + '/' + api_product
        response = self.session.delete(api_product_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        deleted_api_product = response.json()
        assert_valid_schema(deleted_api_product, self.schemas['api_product'])

        return deleted_api_product


    def delete_nonexisting(self, api_product):
        """
        Delete non-existing api product, which should fail
        """
        api_product_url = self.api_product_url + '/' + api_product
        response = self.session.delete(api_product_url)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def assert_compare(self, api_product_a, api_product_b):
        """
        Compares minimum required fields that can be set of two api products
        """
        assert api_product_a['name'] == api_product_b['name']
        assert api_product_a['apiResources'] == api_product_b['apiResources']
        assert api_product_a['approvalType'] == api_product_b['approvalType']
        assert api_product_a['attributes'] == api_product_b['attributes']
        assert api_product_a['status'] == api_product_b['status']
