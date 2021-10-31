"""
APIproduct module does all REST API operations on api_product endpoint
"""
import random
import urllib
from common import assert_status_code, assert_content_type_json, \
                    load_json_schema, assert_valid_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class APIproduct:
    """
    APIproduct does all REST API operations on apiproduct endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.api_product_url = config['api_url'] + '/apiproducts'
        self.schemas = {
            'api-product': load_json_schema('api-product.json'),
            'api-products': load_json_schema('api-products.json'),
            'api-products-names': load_json_schema('api-products-names.json'),
            'error': load_json_schema('error.json'),
        }


    def generate_api_product_name(self, number):
        """
        Returns name of test api product based upon provided number
        """
        return f"testsuite-apiproduct{number}"


    def _create(self, success_expected, new_api_product=None):
        """
        Create new api product to be used as test subject.
        If none provided generate API product with random name.
        """

        if new_api_product is None:
            random_int = random.randint(0,99999)
            new_api_product = {
                "name" : self.generate_api_product_name(random_int),
                "displayName": f'TestSuite product {random_int}',
                "apiResources": [ ],
                "approvalType": "auto",
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.api_product_url, json=new_api_product)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)
            assert_content_type_json(response)

            # Check if just created api product matches with what we requested
            created_api_product = response.json()
            assert_valid_schema(created_api_product, self.schemas['api-product'])
            self.assert_compare(created_api_product, new_api_product)

            return created_api_product

        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['error'])

        return response.json()


    def create_positive(self, new_api_product=None):
        """
        Create new api product, if none provided generate API product with random data.
        """
        return self._create(True, new_api_product)


    def create_negative(self, api_product):
        """
        Attempt to create api product which already exists, should fail
        """
        return self._create(False, api_product)


    def get_all(self):
        """
        Get all api products
        """
        response = self.session.get(self.api_product_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['api-products'])
        # TODO filtering on attributename, attributevalue, startKey

        return response.json()


    def get_all_detailed(self):
        """
        Get all api products with full details
        """
        response = self.session.get(self.api_product_url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['api-products-names'])
        # TODO filtering on attributename, attributevalue, startKey


    def get_positive(self, api_product):
        """
        Get existing api product
        """
        api_product_url = self.api_product_url + '/' + api_product
        response = self.session.get(api_product_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_api_product = response.json()
        assert_valid_schema(retrieved_api_product, self.schemas['api-product'])

        return retrieved_api_product


    def update_positive(self, api_product, updated_api_product):
        """
        Update existing api product
        """
        api_product_url = self.api_product_url + '/' + api_product
        response = self.session.post(api_product_url, json=updated_api_product)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_api_product = response.json()
        assert_valid_schema(updated_api_product, self.schemas['api-product'])

        return updated_api_product


    def _delete(self, api_product_name, expected_success):
        """
        Delete existing api product
        """
        api_product_url = self.api_product_url + '/' + urllib.parse.quote(api_product_name)
        response = self.session.delete(api_product_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
            assert_content_type_json(response)
            assert_valid_schema(response.json(), self.schemas['api-product'])
        else:
            assert_status_code(response, HTTP_NOT_FOUND)
            assert_content_type_json(response)

        return response.json()


    def delete_positive(self, api_product_name):
        """
        Delete existing api product
        """
        return self._delete(api_product_name, True)


    def delete_negative(self, api_product_name):
        """
        Attempt to delete non-existing api product, should fail
        """
        return self._delete(api_product_name, False)


    def delete_negative_bad_request(self, api_product_name):
        """
        Attempt to delete product associated with assigned keys, should fail
        """
        api_product_url = self.api_product_url + '/' + urllib.parse.quote(api_product_name)
        response = self.session.delete(api_product_url)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)


    def assert_compare(self, api_product_a, api_product_b):
        """
        Compares minimum required fields that can be set of two api products
        """
        assert api_product_a['name'] == api_product_b['name']
        assert api_product_a['apiResources'] == api_product_b['apiResources']
        assert api_product_a['approvalType'] == api_product_b['approvalType']
        assert api_product_a['attributes'] == api_product_b['attributes']
