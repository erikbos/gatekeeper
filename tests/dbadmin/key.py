"""
Key module does all REST API operations on key endpoint
"""
import urllib
from common import assert_status_code, assert_status_codes
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, \
    HTTP_CONFLICT, HTTP_NO_CONTENT, HTTP_BAD_REQUEST


class Key:
    """
    Key does all REST API operations on key endpoint
    """

    def __init__(self, config, session, developer_email, app_name):
        self.config = config
        self.session = session
        if developer_email is not None:
            self.url = (config['api_url'] +
                                    '/v1/organizations/' + config['api_organization'] +
                                    '/developers/' + urllib.parse.quote(developer_email) +
                                    '/apps/' + urllib.parse.quote(app_name) + '/keys')


    def _create(self, new_key, success_expected):
        """
        Create new key
        """
        response = self.session.post(self.url + "/create", json=new_key)

        if success_expected:
            assert_status_code(response, HTTP_CREATED)
            # Check if just created key matches with what we requested
            created_key = response.json()
            self.assert_compare(created_key, new_key)

            return created_key

        assert_status_codes(response, [ HTTP_BAD_REQUEST, HTTP_CONFLICT ] )


    def create_positive(self, new_key):
        """
        Create and import new key
        """
        return self._create(new_key, True)


    def create_negative(self, new_key):
        """
        Create and import new key, should fail
        """
        self._create(new_key, False)


    def get_positive(self, consumer_key):
        """
        Get existing key
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(consumer_key))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _update(self, consumer_key, updated_key, success_expected):
        """
        Update existing key
        """
        print ("updated_key", updated_key)

        headers = self.session.headers
        headers['content-type'] = 'application/json'
        key_url = self.url + '/' + urllib.parse.quote(consumer_key)
        response = self.session.post(key_url, headers=headers, json=updated_key)

        print ("res", response.text)
        if success_expected:
            assert_status_code(response, HTTP_OK)
            return response.json()

        assert_status_codes(response, [ HTTP_BAD_REQUEST ])


    def update_positive(self, consumer_key, updated_key):
        """
        Update existing key
        """
        return self._update(consumer_key, updated_key, True)


    def update_negative(self, consumer_key, updated_key):
        """
        Update existing key with changed consumerKey, should fail
        """
        self._update(consumer_key, updated_key, False)


    def _change_status(self, consumer_key, status, expect_success):
        """
        Update key status to a value that is supported
        """
        headers = self.session.headers
        headers['content-type'] = 'application/octet-stream'
        key_url = self.url + '/' + urllib.parse.quote(consumer_key) + '?action=' + status

        response = self.session.post(key_url, headers=headers)

        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def change_status_approve_positive(self, consumer_key):
        """
        Update status of key to approve
        """
        self._change_status(consumer_key, 'approve', True)


    def change_status_revoke_positive(self, consumer_key):
        """
        Update status of key to revoke
        """
        self._change_status(consumer_key, 'revoke', True)


    def _delete(self, consumer_key, expect_success):
        """
        Delete existing key
        """
        response = self.session.delete(self.url + '/' + urllib.parse.quote(consumer_key))
        if expect_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)
        return response.json()


    def delete_positive(self, consumer_key):
        """
        Delete existing key
        """
        return self._delete(consumer_key, True)


    def delete_negative(self, consumer_key):
        """
        Attempt to delete key, which should fail
        """
        return self._delete(consumer_key, False)


    def assert_compare(self, key_a, key_b):
        """
        Compares minimum required fields that can be set of two keys
        """
        assert key_a['consumerKey'] == key_b['consumerKey']
        assert key_a['consumerSecret'] == key_b['consumerSecret']


    def api_product_change_status(self, consumer_key, api_product_name, status, expect_success):
        """
        Update apiproduct status to a value that is supported
        """
        headers = self.session.headers
        headers['content-type'] = 'application/octet-stream'
        key_url = (self.url + '/' + urllib.parse.quote(consumer_key)
                                + '/apiproducts/' + urllib.parse.quote(api_product_name) + '?action=' + status)
        response = self.session.post(key_url, headers=headers)

        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def _api_product_delete(self, consumer_key, api_product_name, expect_success):
        """
        Delete apiproduct from key
        """
        key_api_product_url = (self.url + '/' + urllib.parse.quote(consumer_key)
                                            + '/apiproducts/' + urllib.parse.quote(api_product_name))
        response = self.session.delete(key_api_product_url)
        if expect_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def api_product_delete_positive(self, consumer_key, api_product_name):
        """
        Delete existing key
        """
        return self._api_product_delete(consumer_key, api_product_name, True)


    def api_product_delete_negative(self, consumer_key, api_product_name):
        """
        Attempt to delete key, which should fail
        """
        return self._api_product_delete(consumer_key, api_product_name, False)
