"""
Test suite to validate companies endpoints operations
"""
import copy
import random
import urllib
from common import get_config, get_http_session
from company import Company
from attribute import run_attribute_tests

config = get_config()
session = get_http_session(config)


def test_company_get_all():
    """
    Test get all companies
    """
    company_api = Company(config, session)
    company_api.get_all()


def test_company_get_all_detailed():
    """
    Test get all companies with full details
    """
    company_api = Company(config, session)
    company_api.get_all_detailed()


def test_company_crud_lifecycle():
    """
    Test create, read, update, delete one company
    """
    company_api = Company(config, session)
    created_company = company_api.create_positive()

    # Creating same, now existing company, must not be possible
    company_api.create_negative(created_company)

    # Read existing company by name
    retrieved_company = company_api.get_positive(created_company['name'])
    company_api.assert_compare(retrieved_company, created_company)

    # Check if created company shows up in global company list
    if created_company['name'] not in company_api.get_all():
        assert()

    # Update to new email address and attributes
    updated_company = copy.deepcopy(created_company)
    random_int = random.randint(0,99999)
    updated_company['displayName'] = f'Testsuite Company Name {random_int}'
    updated_company['attributes'] = [
              {
                   "name" : "Test",
              },
              {
                   "name" : "Status",
                   "value" : "Royal"
              }
         ]
    company_api.update_positive(created_company['name'], updated_company)

    # Check if updated company shows up in global company list
    if updated_company['name'] not in company_api.get_all():
        assert()

    # Delete just created company
    deleted_company = company_api.delete_positive(updated_company['name'])
    company_api.assert_compare(deleted_company, updated_company)

    # Try to delete company once more, must not exist anymore
    company_api.delete_negative(updated_company['name'])


def test_company_attributes():
    """
    Test create, read, update, delete attributes of company
    """
    company_api = Company(config, session)

    test_company = company_api.create_positive()

    company_attributes_url = (company_api.company_url +
        '/' + urllib.parse.quote(test_company['name']) + '/attributes')
    run_attribute_tests(config, session, company_attributes_url)

    # clean up
    company_api.delete_positive(test_company['name'])


def test_company_change_status():
    """
    Test changing status of company
    """
    company_api = Company(config, session)
    test_company = company_api.create_positive()

    # Company status should be active after creation
    assert test_company['status'] == 'active'

    # Change status to inactive & active
    name = test_company['name']
    company_api.change_status_inactive_positive(name)
    assert company_api.get_positive(name)['status'] == 'inactive'

    company_api.change_status_active_positive(name)
    assert company_api.get_positive(name)['status'] == 'active'

    company_api.delete_positive(name)
