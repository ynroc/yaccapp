package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/stashapp/stash/internal/manager"
	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/utils"
)

type performerRoutes struct {
	txnManager models.TransactionManager
}

func (rs performerRoutes) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{performerId}", func(r chi.Router) {
		r.Use(PerformerCtx)
		r.Get("/image", rs.Image)
	})

	return r
}

func (rs performerRoutes) Image(w http.ResponseWriter, r *http.Request) {
	performer := r.Context().Value(performerKey).(*models.Performer)
	defaultParam := r.URL.Query().Get("default")

	var image []byte
	if defaultParam != "true" {
		readTxnErr := rs.txnManager.WithReadTxn(r.Context(), func(repo models.ReaderRepository) error {
			image, _ = repo.Performer().GetImage(performer.ID)
			return nil
		})
		if readTxnErr != nil {
			logger.Warnf("couldn't execute getting a performer image from read transaction: %v", readTxnErr)
		}
	}

	if len(image) == 0 || defaultParam == "true" {
		image, _ = getRandomPerformerImageUsingName(performer.Name.String, performer.Gender.String, config.GetInstance().GetCustomPerformerImageLocation())
	}

	if err := utils.ServeImage(image, w, r); err != nil {
		logger.Warnf("error serving image: %v", err)
	}
}

func PerformerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performerID, err := strconv.Atoi(chi.URLParam(r, "performerId"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		var performer *models.Performer
		if err := manager.GetInstance().TxnManager.WithReadTxn(r.Context(), func(repo models.ReaderRepository) error {
			var err error
			performer, err = repo.Performer().Find(performerID)
			return err
		}); err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		ctx := context.WithValue(r.Context(), performerKey, performer)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
