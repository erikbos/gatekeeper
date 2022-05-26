# managementserver tests

This directory contains test suite to test developer endpoints.

## Environment configuration

Environment variable required to be set before running tests:

* API_URL, management server URL (e.g. `https://127.0.0.1:7777/`)
* API_USERNAME, username used for HTTP Basic authentication (default managementserver is 'admin')
* API_PASSWORD, password used for HTTP Basic authentication (default managementserver is 'passwd')
* API_ORGANIZATION, optional, if not set 'default' will be used

## Example

    # Go to directory with all tests
    cd tests/managementserver

    # Set environment with management connection details
    export API_URL=https://127.0.0.1:7777/
    export API_USERNAME=admin
    export API_PASSWORD=passwd

    # Run all developer related tests
    pytest -v test_developer

    # Run five tests in parallel
    pytest -n 5 test_developer

    # Run one test named 'test_crud_lifecycle_10_developers'
    pytest -k test_get_all_developers_list
