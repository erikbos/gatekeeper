"""
Key module does all REST API operations on key endpoint
"""
from common import assert_status_code, assert_content_type_json, \
    assert_valid_schema, assert_valid_schema_error, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_CONFLICT, HTTP_NO_CONTENT, HTTP_BAD_REQUEST


class Key:
    """
    Key does all REST API operations on key endpoint
    """

    def __init__(self, config, session, developer_email, app_name):
        self.config = config
        self.session = session
        if developer_email is not None:
            self.key_url = (config['api_url'] +
                                    '/developers/' + developer_email +
                                    '/apps/' + app_name + '/keys')
        self.schemas = {
            'key': load_json_schema('key.json'),
            'error': load_json_schema('error.json'),
        }


    def create(self, new_key, success_expected):
        """
        Create new key to be used as test subject.
        """

        response = self.session.post(self.key_url + "/create", json=new_key)

        if success_expected:
            assert_status_code(response, HTTP_CREATED)
            assert_content_type_json(response)

            # Check if just created key matches with what we requested
            created_key = response.json()
            assert_valid_schema(created_key, self.schemas['key'])
            self.assert_compare(created_key, new_key)

            return created_key

        assert_status_code(response, HTTP_CONFLICT)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['error'])


    def get_positive(self, consumer_key):
        """
        Get existing key
        """
        response = self.session.get(self.key_url + '/' + consumer_key)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_key = response.json()
        assert_valid_schema(retrieved_key, self.schemas['key'])

        return retrieved_key


    def update_positive(self, consumer_key, updated_key):
        """
        Update existing key
        """
        response = self.session.post(self.key_url + '/' + consumer_key, json=updated_key)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_key = response.json()
        assert_valid_schema(updated_key, self.schemas['key'])

        return updated_key


    def change_status(self, consumer_key, status, expect_success):
        """
        Update key status to a value that is supported
        """
        headers = self.session.headers
        headers['content-type'] = 'application/octet-stream'
        key_url = self.key_url + '/' + consumer_key + '?action=' + status
        response = self.session.post(key_url, headers=headers)

        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)
            assert_valid_schema_error(response.json())


    def delete_positive(self, consumer_key):
        """
        Delete existing key
        """
        response = self.session.delete(self.key_url + '/' + consumer_key)
        print("del key", self.key_url + '/' + consumer_key)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        deleted_key = response.json()
        assert_valid_schema(deleted_key, self.schemas['key'])

        return deleted_key


    def delete_negative(self, consumer_key):
        """
        Delete non-existing key, which should fail
        """
        response = self.session.delete(self.key_url + '/' + consumer_key)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def assert_compare(self, key_a, key_b):
        """
        Compares minimum required fields that can be set of two keys
        """
        assert key_a['consumerKey'] == key_b['consumerKey']
        assert key_a['consumerSecret'] == key_b['consumerSecret']
