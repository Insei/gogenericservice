package services

import (
	"context"
	"strings"

	errors "github.com/insei/gerro"
	repoInterfaces "github.com/insei/gogenericrepo/interfaces"
	mapper "github.com/insei/gomapper"

	"github.com/insei/godtos"
)

//Create add new entity by create dto
//
//Params:
//	ctx - context
//	repo - repository for entity
//	createDto - create dto for entity
//Return:
//	DtoType - created entity dto
//	error - if an error occurs, otherwise nil
func Create[DtoType any, EntityIDType comparable, EntityType repoInterfaces.IEntity[EntityIDType], CreateDtoType any](ctx context.Context,
	repo repoInterfaces.IGenericRepository[EntityIDType, EntityType], createDto CreateDtoType) (DtoType, error) {
	entity := new(EntityType)
	outDto := new(DtoType)
	err := mapper.Map(createDto, entity)
	if err != nil {
		return *outDto, errors.Internal.Wrap(err, "error map entity to dto")
	}
	newEntity, err := repo.Insert(ctx, *entity)
	if err != nil {
		return *outDto, errors.Internal.Wrap(err, "create entity error")
	}
	err = mapper.Map(newEntity, outDto)
	if err != nil {
		return *outDto, errors.Internal.Wrap(err, "error map dto to entity")
	}
	return *outDto, nil
}

//AddSearchInAllFields add search in all field to query builder
func AddSearchInAllFields[EntityIDType comparable, EntityType repoInterfaces.IEntity[EntityIDType]](
	search string,
	repo repoInterfaces.IGenericRepository[EntityIDType, EntityType],
	queryBuilder repoInterfaces.IQueryBuilder,
) {
	entityModel := new(EntityType)
	stringFieldNames := &[]string{}
	getStringFieldsNames(entityModel, stringFieldNames)
	queryGroup := repo.NewQueryBuilder(nil)
	for i := 0; i < len(*stringFieldNames); i++ {
		fieldName := (*stringFieldNames)[i]
		containPass := strings.Contains(strings.ToLower(fieldName), "pass")
		containKey := strings.Contains(strings.ToLower(fieldName), "key")
		if containPass || containKey {
			continue
		}

		queryGroup.Or(fieldName, "LIKE", "%"+search+"%")
	}
	queryBuilder = queryBuilder.WhereQuery(queryGroup)
}

//GetByID Get entity dto by ID
//
//Params:
//	ctx - context
//	repo - repository for entity
//	id - entity id
//	queryBuilder - query builder with filtering, can be nil
//Return:
//	*DtoType - pointer to dto
//	error - if an error occurs, otherwise nil
func GetByID[
	DtoType any,
	EntityIDType comparable,
	EntityType repoInterfaces.IEntity[EntityIDType],
](
	ctx context.Context,
	repo repoInterfaces.IGenericRepository[EntityIDType, EntityType],
	id EntityIDType,
	queryBuilder repoInterfaces.IQueryBuilder,
) (
	DtoType,
	error,
) {
	dto := new(DtoType)
	entity, err := repo.GetByIDExtended(ctx, id, queryBuilder)
	if err != nil {
		return *dto, err
	}
	err = mapper.Map(entity, dto)
	if err != nil {
		return *dto, errors.Internal.Wrap(err, "error map entity to dto")
	}
	return *dto, nil
}

//Update save the changes to the existing entity via update dto
//
//Params:
//	ctx - context is used only for logging
//	repo - repository for entity
//	updateDto - update dto
//	id - entity id
//	queryBuilder - query builder with extended filtering, can be nil
//Return:
//	DtoType - updated entity dto
//	error - if an error occurs, otherwise nil
func Update[
	DtoType,
	UpdateDtoType any,
	EntityIDType comparable,
	EntityType repoInterfaces.IEntity[EntityIDType],
](
	ctx context.Context,
	repo repoInterfaces.IGenericRepository[EntityIDType, EntityType],
	updateDto UpdateDtoType,
	id EntityIDType,
	queryBuilder repoInterfaces.IQueryBuilder,
) (
	DtoType,
	error,
) {
	entity, err := repo.GetByIDExtended(ctx, id, queryBuilder)
	if err != nil {
		return *new(DtoType), err
	}
	err = mapper.Map(updateDto, &entity)
	if err != nil {
		return *new(DtoType), errors.Internal.Wrap(err, "error map dto to entity")
	}
	updEntity, err := repo.Update(ctx, entity)
	if err != nil {
		return *new(DtoType), errors.Internal.Wrap(err, "failed to update entity in repository")
	}
	dto := new(DtoType)
	err = mapper.Map(updEntity, dto)
	if err != nil {
		return *new(DtoType), errors.Internal.Wrap(err, "error map entity to dto")
	}
	return *dto, nil
}

//GetList Get list of entity dto elements with search in all fields and pagination
//
//Params:
//	ctx - context is used only for logging
//	repo - repository for entity
//	search - string for search in entity string fields
//	orderBy - order by entity field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return:
//	godtos.PaginatedItemsDto[DtoType] - pointer to paginated list
//	error - if an error occurs, otherwise nil
func GetList[DtoType any, EntityIDType comparable, EntityType repoInterfaces.IEntity[EntityIDType]](ctx context.Context,
	repo repoInterfaces.IGenericRepository[EntityIDType, EntityType], search, orderBy, orderDirection string,
	page, pageSize int) (godtos.PaginatedItemsDto[DtoType], error) {
	searchQueryBuilder := repo.NewQueryBuilder(ctx)
	if len(search) > 3 {
		AddSearchInAllFields(search, repo, searchQueryBuilder)
	}
	return GetListExtended[DtoType](ctx, repo, searchQueryBuilder, orderBy, orderDirection, page, pageSize)
}

//GetListExtended Get list of entity dto elements with filtering and pagination
//
//Params:
//	ctx - context is used only for logging
//	repo - repository for entity
//	queryBuilder - query builder with filtering, can be nil
//	orderBy - order by entity field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return:
//	godtos.PaginatedItemsDto[DtoType] - pointer to paginated list
//	error - if an error occurs, otherwise nil
func GetListExtended[
	DtoType any,
	EntityIDType comparable,
	EntityType repoInterfaces.IEntity[EntityIDType],
](
	ctx context.Context,
	repo repoInterfaces.IGenericRepository[EntityIDType, EntityType],
	queryBuilder repoInterfaces.IQueryBuilder,
	orderBy,
	orderDirection string,
	page,
	pageSize int,
) (
	godtos.PaginatedItemsDto[DtoType],
	error,
) {
	paginatedItemsDto := godtos.NewEmptyPaginatedItemsDto[DtoType]()
	pageFinal := page
	pageSizeFinal := pageSize
	if page < 1 {
		pageFinal = 1
	}
	if pageSize < 1 {
		pageSizeFinal = 10
	}
	entities, err := repo.GetList(ctx, orderBy, orderDirection, pageFinal, pageSizeFinal, queryBuilder)
	if err != nil {
		return paginatedItemsDto, errors.Internal.Wrap(err, "repository failed get list")
	}
	count, err := repo.Count(ctx, queryBuilder)
	if err != nil {
		return paginatedItemsDto, errors.Internal.Wrap(err, "error counting entities")
	}
	dtoArr := &[]DtoType{}
	for i := 0; i < len(entities); i++ {
		dto := new(DtoType)
		err = mapper.Map((entities)[i], dto)
		if err != nil {
			return paginatedItemsDto, errors.Internal.Wrap(err, "error map entity to dto")
		}
		*dtoArr = append(*dtoArr, *dto)
	}

	paginatedDto := godtos.NewPaginatedItemsDto[DtoType](pageFinal, pageSizeFinal, count, *dtoArr)
	return paginatedDto, nil
}
