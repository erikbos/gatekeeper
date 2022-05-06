"""
Application module does all REST API operations on application endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST, HTTP_NO_CONTENT


class Application:
    """
    Application does all REST API functions on application endpoint
    """
    def __init__(self, config, session, developer_email):
        self.config = config
        self.session = session
        self.global_application_url = self.config['api_url'] + '/apps'
        if developer_email is not None:
            self.application_url = config['api_url'] + '/developers/' + urllib.parse.quote(developer_email) + '/apps'


    def generate_app_name(self, number):
        """
        Returns generated test application name
        """
        return f"testsuite-app{number}"


    def _create(self, success_expected, new_application=None):
        """
        Create new application to be used as test subject.
        If no app data provided generate application with random data
        """
        if new_application is None:
            random_int = random.randint(0,99999)
            new_application = {
                "name" : self.generate_app_name(random_int),
                # "displayName": "Testsuite app",
                "callbackUrl": "1",
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.application_url, json=new_application)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)
            # Check if just created application matches with what we requested
            self.assert_compare(response.json(), new_application)
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)

        return response.json()


    def create_new(self, new_application=None):
        """
        Create new application, if no app data provided generate application with random data
        """
        return self._create(True, new_application)


    def create_negative(self, application):
        """
        Attempt to create new application which already exists, should fail
        """
        return self._create(False, application)


    def create_key_positive(self, app_name, api_product_name):
        """
        Create new key with assigned api product
        """
        new_key = {
            "apiProducts": [ api_product_name ],
            "status": "approved"
        }

        headers = self.session.headers
        headers['content-type'] = 'application/json'
        application_url = self.application_url + '/' + urllib.parse.quote(app_name)
        response = self.session.put(application_url, headers=headers, json=new_key)
        assert_status_code(response, HTTP_OK)

        return response.json()


    def get_all_global(self):
        """
        Get global list of all application uuids
        """
        response = self.session.get(self.global_application_url)
        assert_status_code(response, HTTP_OK)

        return response.json()


    def get_all_global_detailed(self):
        """
        Get global list of all application names
        """
        response = self.session.get(self.global_application_url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        # TODO testing of paginating response
        # TODO filtering of apptype, expand, rows, startKey, status queryparameters to filter


    def get_all(self):
        """
        Get all application names of one developer
        """
        response = self.session.get(self.application_url)
        assert_status_code(response, HTTP_OK)

        return response.json()


    def get_positive(self, app_name):
        """
        Get existing application
        """
        application_url = self.application_url + '/' + urllib.parse.quote(app_name)
        response = self.session.get(application_url)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def get_by_uuid_positive(self, app_uuid):
        """
        Get existing application by uuid
        """
        application_url = self.global_application_url + '/' + urllib.parse.quote(app_uuid)
        response = self.session.get(application_url)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, application):
        """
        Update existing application
        """
        application_url = self.application_url + '/' + urllib.parse.quote(application['name'])
        response = self.session.post(application_url, json=application)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _change_status(self, app_name, status, expect_success):
        """
        Update status of application
        """
        headers = self.session.headers
        headers['content-type'] = 'application/octet-stream'
        application_url = (self.application_url + '/' +
                    urllib.parse.quote(app_name) + '?action=' + status)
        response = self.session.post(application_url, headers=headers)

        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def change_status_approve_positive(self, app_name):
        """
        Update status of developer to inactive
        """
        self._change_status(app_name, 'approve', True)


    def change_status_revoke_positive(self, app_name):
        """
        Update status of developer to active
        """
        self._change_status(app_name, 'revoke', True)


    def _delete(self, app_name, expected_success):
        """
        Delete existing application
        """
        response = self.session.delete(self.application_url + '/' + urllib.parse.quote(app_name))
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)
        return response.json()


    def delete_positive(self, app_name):
        """
        Delete existing application
        """
        return self._delete(app_name, True)


    def delete_negative(self, app_name):
        """
        Attempt to delete application, which should fail
        """
        return self._delete(app_name, False)


    def delete_all_test_developer(self):
        """
        Delete all application created by test suite
        """
        for i in range(self.config['entity_count']):
            app_name = self.generate_app_name(i)
            application_url = self.application_url + '/' + urllib.parse.quote(app_name)
            self.session.delete(application_url)


    def assert_compare(self, application_a, application_b):
        """
        Compares minimum required fields that can be freely, as is, on applications
        """
        assert application_a['name'] == application_b['name']
        # assert application_a['displayName'] == application_b['displayName']
        assert application_a['callbackUrl'] == application_b['callbackUrl']
        assert (application_a['attributes'].sort(key=lambda x: x['name'])
            == application_b['attributes'].sort(key=lambda x: x['name']))
