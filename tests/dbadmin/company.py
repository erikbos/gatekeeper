"""
Company module does all REST API operations on company endpoint
"""
import random
import urllib
from common import assert_status_code, assert_content_type_json, \
                    load_json_schema, assert_valid_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Company:
    """
    Company does all REST API operations on companies endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.company_url = config['api_url'] + '/companies'
        self.schemas = {
            'company': load_json_schema('company.json'),
            'companies': load_json_schema('companies.json'),
            'company-names': load_json_schema('company-names.json'),
            'error': load_json_schema('error.json'),
        }


    def generate_company_name(self, number):
        """
        Returns name of test company based upon provided number
        """
        return f"testsuite-company{number}"


    def _create(self, success_expected, new_company=None):
        """
        Create new company to be used as test subject.
        If none provided generate API product with random name.
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

        response = self.session.post(self.company_url, json=new_company)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)
            assert_content_type_json(response)

            # Check if just created company matches with what we requested
            created_company = response.json()
            assert_valid_schema(created_company, self.schemas['company'])
            self.assert_compare(created_company, new_company)

            return created_company

        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['error'])

        return response.json()


    def create_positive(self, new_company=None):
        """
        Create new company, if none provided generate API product with random data.
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
        response = self.session.get(self.company_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['company-names'])
        # TODO filtering on attributename, attributevalue, startKey

        return response.json()


    def get_all_detailed(self):
        """
        Get all companies with full details
        """
        response = self.session.get(self.company_url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['companies'])
        # TODO filtering on attributename, attributevalue, startKey


    def get_positive(self, company):
        """
        Get existing company
        """
        company_url = self.company_url + '/' + urllib.parse.quote(company)
        response = self.session.get(company_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_company = response.json()
        assert_valid_schema(retrieved_company, self.schemas['company'])

        return retrieved_company


    def update_positive(self, company, updated_company):
        """
        Update existing company
        """
        company_url = self.company_url + '/' + urllib.parse.quote(company)
        response = self.session.post(company_url, json=updated_company)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_company = response.json()
        assert_valid_schema(updated_company, self.schemas['company'])

        return updated_company


    def _delete(self, company_name, expected_success):
        """
        Delete existing company
        """
        company_url = self.company_url + '/' + urllib.parse.quote(company_name)
        response = self.session.delete(company_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
            assert_content_type_json(response)
            assert_valid_schema(response.json(), self.schemas['company'])
        else:
            assert_status_code(response, HTTP_NOT_FOUND)
            assert_content_type_json(response)

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


    def delete_negative_bad_request(self, company_name):
        """
        Attempt to delete product associated with assigned keys, should fail
        """
        company_url = self.company_url + '/' + urllib.parse.quote(company_name)
        response = self.session.delete(company_url)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)


    def assert_compare(self, company_a, company_b):
        """
        Compares minimum required fields that can be set of two companies
        """
        assert company_a['name'] == company_b['name']
        assert company_a['status'] == company_b['status']
        assert (company_a['attributes'].sort(key=lambda x: x['name'])
            == company_b['attributes'].sort(key=lambda x: x['name']))
