package handlers

import "github.com/labstack/echo/v4"

var (
	isError       bool = false
	fromProtected bool = false
)

func SetupRoutes(e *echo.Echo, ah *AuthHandler, ch *ContainerHandler, ih *ImageHandler, anh *AnalyticsHandler) {
	e.GET("/login", ah.loginHandler)
	e.POST("/login", ah.loginHandler)

	protectedGroup := e.Group("/", ah.authMiddleware)

	protectedGroup.GET("password", ah.passwordHandler)
	protectedGroup.POST("password", ah.passwordHandler)

	protectedGroup.GET("", ch.listContainers)          // Gets HTML page
	protectedGroup.GET("containers", ch.getContainers) // Gets JSON data
	protectedGroup.GET("containers/:id", ch.getContainer)
	protectedGroup.POST("containers/:id/restart", ch.restartContainer)
	protectedGroup.POST("containers/:id/pause", ch.pauseContainer)
	protectedGroup.POST("containers/:id/unpause", ch.unpauseContainer)
	protectedGroup.GET("containers/:id/status", ch.getContainerStatus)
	protectedGroup.GET("containers/:id/logs", ch.getContainerLogs)
	protectedGroup.POST("containers/:id/stop", ch.stopContainer)
	protectedGroup.POST("containers/:id/start", ch.startContainer)
	protectedGroup.POST("containers/:id/recreate", ch.recreateContainer)
	protectedGroup.DELETE("containers/:id/remove", ch.removeContainer)
	protectedGroup.GET("containers/:id/stats", ch.getContainerStats)
	protectedGroup.POST("mysql/create", ch.setupMySQLContainer)
	protectedGroup.POST("containers/:id/exec", ch.execCommandInContainer)
	protectedGroup.POST("prune/containers", ch.pruneContainers)
	protectedGroup.POST("prune/images", ch.pruneImages)
	protectedGroup.POST("prune/volumes", ch.pruneVolumes)
	protectedGroup.POST("prune/networks", ch.pruneNetworks)
	protectedGroup.POST("prune/all", ch.pruneAll)

	protectedGroup.GET("images", ih.listImages)
	protectedGroup.GET("images/fetch", ih.getImages)
	protectedGroup.POST("images/pull", ih.pullImage)
	protectedGroup.DELETE("images/:id/remove", ih.removeImage)
	protectedGroup.POST("images/:id/container", ih.createContainer)
	protectedGroup.GET("images/:id/containers", ih.getUsedContainers)

	protectedGroup.GET("analytics", anh.getAnalyticsPage)
	protectedGroup.GET("analytics/usage", anh.getUsage)
}
