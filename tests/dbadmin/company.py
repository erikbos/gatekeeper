"""
Company module does all REST API operations on company endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST, HTTP_NO_CONTENT


class Company:
    """
    Company does all REST API operations on company endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/v1/organizations/' + config['api_organization'] + '/companies'


    def generate_company_name(self, number):
        """
        Returns name of test company based upon provided number
        """
        return f"testsuite-company{number}"


    def _create(self, success_expected, new_company=None):
        """
        Create new company to be used as test subject.
        If none provided generate company with random name.
        """

        if new_company is None:
            random_int = random.randint(0,99999)
            new_company = {
                "name" : self.generate_company_name(random_int),
                "displayName": f'TestSuite Company {random_int}',
                "status": "active",
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.url, json=new_company)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)

            # Check if just created company matches with what we requested
            created_company = response.json()
            self.assert_compare(created_company, new_company)

            return created_company

        assert_status_code(response, HTTP_BAD_REQUEST)
        return response.json()


    def create_positive(self, new_company=None):
        """
        Create new company, if none provided generate company with random data.
        """
        return self._create(True, new_company)


    def create_negative(self, company):
        """
        Attempt to create company which already exists, should fail
        """
        return self._create(False, company)


    def get_all(self):
        """
        Get all companies
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)
        # TODO filtering on attributename, attributevalue, startKey

        return response.json()


    def get_all_detailed(self):
        """
        Get all companies with full details
        """
        response = self.session.get(self.url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        # TODO filtering on attributename, attributevalue, startKey


    def get_positive(self, company):
        """
        Get existing company
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(company))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, company, updated_company):
        """
        Update existing company
        """
        response = self.session.post(self.url + '/' + urllib.parse.quote(company), json=updated_company)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _change_status(self, company_name, status, expect_success):
        """
        Update status of company
        """
        response = self.session.change_status(self.url + '/' + urllib.parse.quote(company_name), status)
        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def change_status_active_positive(self, company_name):
        """
        Update status of company to active
        """
        self._change_status(company_name, 'active', True)


    def change_status_inactive_positive(self, company_name):
        """
        Update status of company to inactive
        """
        self._change_status(company_name, 'inactive', True)


    def _delete(self, company_name, expected_success):
        """
        Delete existing company
        """
        company_url = self.url + '/' + urllib.parse.quote(company_name)
        response = self.session.delete(company_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)

        return response.json()


    def delete_positive(self, company_name):
        """
        Delete existing company
        """
        return self._delete(company_name, True)


    def delete_negative(self, company_name):
        """
        Attempt to delete non-existing company, should fail
        """
        return self._delete(company_name, False)


    def assert_compare(self, company_a, company_b):
        """
        Compares minimum required fields that can be set of two companies
        """
        assert company_a['name'] == company_b['name']
        assert company_a['status'] == company_b['status']
        assert (company_a['attributes'].sort(key=lambda x: x['name'])
            == company_b['attributes'].sort(key=lambda x: x['name']))
