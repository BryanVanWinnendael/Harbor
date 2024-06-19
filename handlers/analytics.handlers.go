package handlers

import (
	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/BryanVanWinnendael/Harbor/views/analytics_views"
	"github.com/labstack/echo/v4"
)

type AnalyticsService interface {
	GetTotalUsage() (dto.UsageDTO, error)
}

func NewAnalyticsHandler(as AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		AnalyticsServices: as,
	}
}

type AnalyticsHandler struct {
	AnalyticsServices AnalyticsService
}

func (is *AnalyticsHandler) getAnalyticsPage(c echo.Context) error {
	return renderView(c, analytics_views.AnalyticsIndex(
		"Analytics |",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		analytics_views.AnalyticsHome(),
	))
}

func (is *AnalyticsHandler) getUsage(c echo.Context) error {
	usage, err := is.AnalyticsServices.GetTotalUsage()
	if err != nil {
		return c.JSON(500, map[string]string{
			"error": err.Error(),
		})
	}

	return renderView(c, analytics_views.AnalyticsUsage(usage))
}
