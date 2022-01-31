package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// CompanyService is
type CompanyService struct {
	db    *db.Database
	audit *audit.Audit
}

// NewCompany returns a new company instance
func NewCompany(database *db.Database, a *audit.Audit) *CompanyService {

	return &CompanyService{
		db:    database,
		audit: a,
	}
}

// GetAll returns all companies
func (ds *CompanyService) GetAll(organizationName string) (companys types.Companies, err types.Error) {

	return ds.db.Company.GetAll(organizationName)
}

// Get returns details of a company
func (ds *CompanyService) Get(organizationName, companyName string) (company *types.Company, err types.Error) {

	return ds.db.Company.Get(organizationName, companyName)
}

// Create creates a new company
func (ds *CompanyService) Create(organizationName string, newCompany types.Company,
	who audit.Requester) (*types.Company, types.Error) {

	if _, err := ds.Get(organizationName, newCompany.Name); err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("company '%s' already exists", newCompany.Name))
	}
	// Automatically set default fields
	newCompany.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newCompany.CreatedBy = who.User
	newCompany.Activate()

	if err := ds.updateCompany(organizationName, &newCompany, who); err != nil {
		return nil, err
	}
	env := &audit.Environment{
		Organization: organizationName,
		Company:      newCompany.Name,
	}
	ds.audit.Create(newCompany, env, who)
	return &newCompany, nil
}

// Update updates an existing company
func (ds *CompanyService) Update(organizationName string, updatedCompany types.Company,
	who audit.Requester) (*types.Company, types.Error) {

	currentCompany, err := ds.db.Company.Get(organizationName, updatedCompany.Name)
	if err != nil {
		return nil, err
	}

	// Copy over fields we do not allow to be updated
	updatedCompany.Name = currentCompany.Name
	updatedCompany.CreatedAt = currentCompany.CreatedAt
	updatedCompany.CreatedBy = currentCompany.CreatedBy

	if err = ds.updateCompany(organizationName, &updatedCompany, who); err != nil {
		return nil, err
	}
	env := &audit.Environment{
		Organization: organizationName,
		Company:      updatedCompany.Name,
	}
	ds.audit.Update(currentCompany, updatedCompany, env, who)
	return &updatedCompany, nil
}

// updateCompany updates last-modified field(s) and updates company in database
func (ds *CompanyService) updateCompany(organizationName string,
	updatedCompany *types.Company, who audit.Requester) types.Error {

	updatedCompany.Attributes.Tidy()
	updatedCompany.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedCompany.LastModifiedBy = who.User

	if err := updatedCompany.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return ds.db.Company.Update(organizationName, updatedCompany)
}

// Delete deletes an company
func (ds *CompanyService) Delete(organizationName, companyName string,
	who audit.Requester) (e types.Error) {

	company, err := ds.Get(organizationName, companyName)
	if err != nil {
		return err
	}
	// FIXME
	// keyWithCompany, err := ds.db.Key.GetCountByCompanyName(organizationName, companyName)
	// if err != nil {
	// 	return err
	// }
	// if keyWithCompany > 0 {
	// 	return types.NewBadRequestError(
	// 		fmt.Errorf("cannot delete api product '%s' assigned to %d keys",
	// 			companyName, keyWithCompany))
	// }
	if err := ds.db.Company.Delete(organizationName, companyName); err != nil {
		return err
	}
	env := &audit.Environment{
		Organization: organizationName,
		Company:      companyName,
	}
	ds.audit.Delete(company, env, who)
	return nil
}
