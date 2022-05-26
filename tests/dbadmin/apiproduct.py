"""
APIproduct module does all REST API operations on api_product endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class APIproduct:
    """
    APIproduct does all REST API operations on apiproduct endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/v1/organizations/' + config['api_organization'] + '/apiproducts'


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
                "scopes" : [
                     f"scope{random_int}",
                     f"another_scope{random_int}",
                    ],
            }

        response = self.session.post(self.url, json=new_api_product)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)

            # Check if just created api product matches with what we requested
            created_api_product = response.json()
            self.assert_compare(created_api_product, new_api_product)

            return created_api_product

        assert_status_code(response, HTTP_BAD_REQUEST)
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
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)
        # TODO filtering on attributename, attributevalue, startKey

        return response.json()


    def get_all_detailed(self):
        """
        Get all api products with full details
        """
        response = self.session.get(self.url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        # TODO filtering on attributename, attributevalue, startKey


    def get_positive(self, api_product):
        """
        Get existing api product
        """
        api_product_url = self.url + '/' + urllib.parse.quote(api_product)
        response = self.session.get(api_product_url)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, api_product, updated_api_product):
        """
        Update existing api product
        """
        api_product_url = self.url + '/' + urllib.parse.quote(api_product)
        response = self.session.post(api_product_url, json=updated_api_product)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _delete(self, api_product_name, expected_success):
        """
        Delete existing api product
        """
        api_product_url = self.url + '/' + urllib.parse.quote(api_product_name)
        response = self.session.delete(api_product_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)

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
        api_product_url = self.url + '/' + urllib.parse.quote(api_product_name)
        response = self.session.delete(api_product_url)
        assert_status_code(response, HTTP_BAD_REQUEST)


    def assert_compare(self, api_product_a, api_product_b):
        """
        Compares minimum required fields that can be set of two api products
        """
        assert api_product_a['name'] == api_product_b['name']
        assert api_product_a['apiResources'] == api_product_b['apiResources']
        assert api_product_a['approvalType'] == api_product_b['approvalType']
        assert (api_product_a['attributes'].sort(key=lambda x: x['name'])
            == api_product_b['attributes'].sort(key=lambda x: x['name']))
