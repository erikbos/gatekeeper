"""
Developer module does all REST API operations on developer endpoint
"""
import random
import urllib
from common import assert_status_code, assert_status_codes, assert_content_type_json
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST, HTTP_NO_CONTENT


class Developer:
    """
    Developer does all REST API operations on developer endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/developers'


    def generate_email_address(self, number):
        """
        Returns email address of test developer based upon provided number
        """
        return f"testsuite-dev{number}@test.com"


    def _create(self, success_expected, new_developer=None):
        """
        Create new developer, if none provided generate random developer
        """
        if new_developer is None:
            random_int = random.randint(0,99999)
            new_developer = {
                "email" : self.generate_email_address(random_int),
                "firstName" : f"first{random_int}",
                "lastName" : f"last{random_int}",
                "userName" : f"username{random_int}",
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        # Workaround for intermittent 415s:
        #   request module should take of content-type given we provide json
        #   or perhaps it sets wrong content-type..
        headers = self.session.headers
        headers['content-type'] = 'application/json'
        response = self.session.post(self.url, headers=headers, json=new_developer)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)
            assert_content_type_json(response)

            # Check if just created developer matches with what we requested
            self.assert_compare(response.json(), new_developer)
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)

        return response.json()

    def create_positive(self, new_developer=None):
        """
        Create new developer to be used as test subject. If none provided generate random developer
        """
        return self._create(True, new_developer)


    def create_negative(self, developer):
        """
        Attempt to create developer which already exists, should fail
        """
        return self._create(False, developer)


    def get_all(self):
        """
        Get all developers email addresses
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)
        # TODO testing of paginating response
        # TODO filtering on apptype, expand, rows, startKey, status queryparameters to filter

        return response.json()


    def get_all_detailed(self):
        """
        Get all developers with full details
        """
        response = self.session.get(self.url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        # TODO testing of paginating response
        # TODO filtering on apptype, expand, rows, startKey, status queryparameters to filter


    def get_positive(self, developer_email):
        """
        Get existing developer
        """
        url = self.url + '/' + urllib.parse.quote(developer_email)
        response = self.session.get(url)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, developer_email, developer):
        """
        Update existing developer
        """
        url = self.url + '/' + urllib.parse.quote(developer_email)
        response = self.session.post(url, json=developer)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _change_status(self, developer_email, status, expect_success):
        """
        Update status of developer
        """
        headers = self.session.headers
        headers['content-type'] = 'application/octet-stream'
        url = self.url + '/' + urllib.parse.quote(developer_email) + '?action=' + status
        response = self.session.post(url, headers=headers)

        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def change_status_active_positive(self, developer_email):
        """
        Update status of developer to active
        """
        self._change_status(developer_email, 'active', True)


    def change_status_inactive_positive(self, developer_email):
        """
        Update status of developer to inactive
        """
        self._change_status(developer_email, 'inactive', True)


    def _delete(self, developer_email, expected_success):
        """
        Delete developer
        """
        url = self.url + '/' + urllib.parse.quote(developer_email)
        response = self.session.delete(url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_codes(response, [ HTTP_NOT_FOUND, HTTP_BAD_REQUEST ])
        return response.json()


    def delete_positive(self, developer_email):
        """
        Delete existing developer
        """
        return self._delete(developer_email, True)


    def delete_negative_notfound(self, developer_email):
        """
        Attempt to delete developer, which should fail
        """
        return self._delete(developer_email, False)


    def delete_negative_badrequest(self, developer_email):
        """
        Attempt to delete developer with at least one app assigned, which should fail
        """
        return self._delete(developer_email, False)


    def delete_all_test_developer(self):
        """
        Delete all developers created by test suite
        """
        for i in range(self.config['entity_count']):
            email = self.generate_email_address(i)
            url = self.url + '/' + urllib.parse.quote(email)
            self.session.delete(url)


    def assert_compare(self, developer_a, developer_b):
        """
        Compares minimum required fields that can be set of two developers
        """
        assert developer_a['email'] == developer_b['email']
        assert developer_a['firstName'] == developer_b['firstName']
        assert developer_a['lastName'] == developer_b['lastName']

        assert (developer_a['attributes'].sort(key=lambda x: x['name'])
            == developer_b['attributes'].sort(key=lambda x: x['name']))
