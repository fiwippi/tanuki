package tanuki

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/pretty"
)

func createRouter(efs fs.FS) *gin.Engine {
	// Create the router
	r := gin.New()

	// Attach the middleware
	r.Use(logMiddleware())
	r.Use(gin.Recovery())

	//
	r.MaxMultipartMemory = int64(conf.MaxUploadedFileSizeMiB) << 20

	// Serve static files
	staticFS, err := fs.Sub(efs, static)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create static filesystem")
	}
	r.StaticFS("/static", http.FS(staticFS))

	// Setup the template renderer
	r.HTMLRender = templateRenderer(efs)
	log.Info().Str("templates", pretty.MapKeys(r.HTMLRender)).Msg("templates loaded")

	// Handle 404s
	r.NoRoute(err404)

	// Register routes
	authorised := r.Group("/")
	authorised.Use(authMiddleware())
	setupFrontendRoutes(r, authorised)
	setupAPIRoutes(r, authorised)
	setupOPDSRoutes(r)

	return r
}

func setupFrontendRoutes(r *gin.Engine, authorised *gin.RouterGroup) {
	// If already authorised then skips these routes
	loginGroup := r.Group("/")
	loginGroup.Use(skipIfAuthedMiddleware())
	loginGroup.GET("/login", loginPage)

	// Must be authorised to access these routes
	authorised.GET("/", homePage)
	authorised.GET("/tags", tagsPage)
	authorised.GET("/tags/:tag", specificTagPage)
	authorised.GET("/entries/:sid", entriesPage)
	authorised.GET("/reader/:sid/:eid", readerPage)

	// Must be authorised and an admin to access these routes i.e. /admin
	admin := authorised.Group("/admin")
	admin.Use(adminMiddleware())
	admin.GET("/", adminDashboard)
	admin.GET("/users", adminUsers)
	admin.GET("/users/edit", adminUsersEdit)
	admin.GET("/users/create", adminUsersCreate)
	admin.GET("/missing-items", adminMissingEntries)
}

func setupAPIRoutes(r *gin.Engine, authorised *gin.RouterGroup) {
	r.POST("/api/auth/login", authLogin)
	authorised.GET("/api/auth/logout", authLogout)

	api := authorised.Group("/api")
	api.GET("/tags", apiGetAllTags)
	api.GET("/tag/:tag", apiGetSeriesWithTag)
	api.GET("/series", apiGetCatalog)
	api.GET("/series/:sid", apiGetSeries)
	api.PATCH("/series/:sid", apiPatchSeries)
	api.GET("/series/:sid/cover", apiGetSeriesCover)
	api.GET("/series/:sid/progress", apiGetSeriesProgress)
	api.PATCH("/series/:sid/progress", apiPatchSeriesProgress)
	api.GET("/catalog/progress", apiGetCatalogProgress)
	api.GET("/series/:sid/entries/:eid/progress", apiGetEntryProgress)
	api.PATCH("/series/:sid/entries/:eid/progress", apiPatchEntryProgress)
	api.PATCH("/series/:sid/cover", apiPatchSeriesCover)
	api.DELETE("/series/:sid/cover", apiDeleteSeriesCover)
	api.GET("/series/:sid/tags", apiGetSeriesTags)
	api.PATCH("/series/:sid/tags", apiPatchSeriesTags)
	api.GET("/series/:sid/entries", apiGetSeriesEntries)
	api.GET("/series/:sid/entries/:eid", apiGetEntry)
	api.PATCH("/series/:sid/entries/:eid", apiPatchEntry)
	api.GET("/series/:sid/entries/:eid/cover", apiGetEntryCover)
	api.PATCH("/series/:sid/entries/:eid/cover", apiPatchEntryCover)
	api.DELETE("/series/:sid/entries/:eid/cover", apiDeleteEntryCover)
	api.GET("/series/:sid/entries/:eid/archive", apiGetEntryArchive)
	api.GET("/series/:sid/entries/:eid/page/:num", apiGetEntryPage)

	// Users can request/edit data about themselves if they
	// provide their cookie to identify themselves
	apiUser := api.Group("/user")
	apiUser.GET("/type", apiGetUserType)
	apiUser.GET("/name", apiGetUserName)

	// Must be an admin to access these api routes i.e. /api/admin/...
	apiAdmin := api.Group("/admin")
	apiAdmin.Use(adminMiddleware())
	apiAdmin.GET("/library/scan", apiGetAdminLibraryScan)
	apiAdmin.GET("/library/generate-thumbnails", apiGetAdminLibraryGenerateThumbnails)
	apiAdmin.GET("/library/missing-items", apiGetAdminLibraryMissingEntries)
	apiAdmin.DELETE("/library/missing-items", apiDeleteAdminLibraryMissingEntries)
	apiAdmin.GET("/db", apiGetAdminDB)
	apiAdmin.GET("/users", apiGetAdminUsers)
	apiAdmin.PUT("/users", apiPutAdminUsers)
	apiAdmin.GET("/user/:id", apiGetAdminUser)
	apiAdmin.PATCH("/user/:id", apiPatchAdminUser)
	apiAdmin.DELETE("/user/:id", apiDeleteAdminUser)

}

func setupOPDSRoutes(r *gin.Engine) {
	opds := r.Group("/opds")
	opds.Use(basicAuthMiddleware("Tanuki OPDS"))

	v1p2 := opds.Group("/v1.2")
	v1p2.GET("catalog", opdsCatalog)
	v1p2.GET("/series/:sid", opdsViewEntries)
	v1p2.GET("/series/:sid/entries/:eid/archive", opdsArchive)
	v1p2.GET("/series/:sid/entries/:eid/cover", opdsCover)
}
