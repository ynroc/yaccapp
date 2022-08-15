package manager

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/stashapp/stash/internal/desktop"
	"github.com/stashapp/stash/internal/dlna"
	"github.com/stashapp/stash/internal/log"
	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/database"
	"github.com/stashapp/stash/pkg/ffmpeg"
	"github.com/stashapp/stash/pkg/fsutil"
	"github.com/stashapp/stash/pkg/job"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/paths"
	"github.com/stashapp/stash/pkg/plugin"
	"github.com/stashapp/stash/pkg/scraper"
	"github.com/stashapp/stash/pkg/session"
	"github.com/stashapp/stash/pkg/sqlite"
	"github.com/stashapp/stash/pkg/utils"
	"github.com/stashapp/stash/ui"
)

type Manager struct {
	Config *config.Instance
	Logger *log.Logger

	Paths *paths.Paths

	FFMPEG  ffmpeg.FFMpeg
	FFProbe ffmpeg.FFProbe

	ReadLockManager *fsutil.ReadLockManager

	SessionStore *session.Store

	JobManager *job.Manager

	PluginCache  *plugin.Cache
	ScraperCache *scraper.Cache

	DownloadStore *DownloadStore

	DLNAService *dlna.Service

	TxnManager models.TransactionManager

	scanSubs *subscriptionManager
}

var instance *Manager
var once sync.Once

func GetInstance() *Manager {
	if _, err := Initialize(); err != nil {
		panic(err)
	}
	return instance
}

func Initialize() (*Manager, error) {
	var err error
	once.Do(func() {
		err = initialize()
	})

	return instance, err
}

func initialize() error {
	ctx := context.TODO()
	cfg, err := config.Initialize()

	if err != nil {
		return fmt.Errorf("initializing configuration: %w", err)
	}

	l := initLog()
	initProfiling(cfg.GetCPUProfilePath())

	instance = &Manager{
		Config:          cfg,
		Logger:          l,
		ReadLockManager: fsutil.NewReadLockManager(),
		DownloadStore:   NewDownloadStore(),
		PluginCache:     plugin.NewCache(cfg),

		TxnManager: sqlite.NewTransactionManager(),

		scanSubs: &subscriptionManager{},
	}

	instance.JobManager = initJobManager()

	sceneServer := SceneServer{
		TXNManager: instance.TxnManager,
	}
	instance.DLNAService = dlna.NewService(instance.TxnManager, instance.Config, &sceneServer)

	if !cfg.IsNewSystem() {
		logger.Infof("using config file: %s", cfg.GetConfigFile())

		if err == nil {
			err = cfg.Validate()
		}

		if err != nil {
			return fmt.Errorf("error initializing configuration: %w", err)
		} else if err := instance.PostInit(ctx); err != nil {
			return err
		}

		initSecurity(cfg)
	} else {
		cfgFile := cfg.GetConfigFile()
		if cfgFile != "" {
			cfgFile += " "
		}

		// create temporary session store - this will be re-initialised
		// after config is complete
		instance.SessionStore = session.NewStore(cfg)

		logger.Warnf("config file %snot found. Assuming new system...", cfgFile)
	}

	if err = initFFMPEG(ctx); err != nil {
		logger.Warnf("could not initialize FFMPEG subsystem: %v", err)
	}

	// if DLNA is enabled, start it now
	if instance.Config.GetDLNADefaultEnabled() {
		if err := instance.DLNAService.Start(nil); err != nil {
			logger.Warnf("could not start DLNA service: %v", err)
		}
	}

	return nil
}

func initJobManager() *job.Manager {
	ret := job.NewManager()

	// desktop notifications
	ctx := context.Background()
	c := ret.Subscribe(context.Background())
	go func() {
		for {
			select {
			case j := <-c.RemovedJob:
				if instance.Config.GetNotificationsEnabled() {
					cleanDesc := strings.TrimRight(j.Description, ".")

					if j.StartTime == nil {
						// Task was never started
						return
					}

					timeElapsed := j.EndTime.Sub(*j.StartTime)
					desktop.SendNotification("Task Finished", "Task \""+cleanDesc+"\" is finished in "+formatDuration(timeElapsed)+".")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ret
}

func formatDuration(t time.Duration) string {
	return fmt.Sprintf("%02.f:%02.f:%02.f", t.Hours(), t.Minutes(), t.Seconds())
}

func initSecurity(cfg *config.Instance) {
	if err := session.CheckExternalAccessTripwire(cfg); err != nil {
		session.LogExternalAccessError(*err)
	}
}

func initProfiling(cpuProfilePath string) {
	if cpuProfilePath == "" {
		return
	}

	f, err := os.Create(cpuProfilePath)
	if err != nil {
		logger.Fatalf("unable to create cpu profile file: %s", err.Error())
	}

	logger.Infof("profiling to %s", cpuProfilePath)

	// StopCPUProfile is defer called in main
	if err = pprof.StartCPUProfile(f); err != nil {
		logger.Warnf("could not start CPU profiling: %v", err)
	}
}

func initFFMPEG(ctx context.Context) error {
	// only do this if we have a config file set
	if instance.Config.GetConfigFile() != "" {
		// use same directory as config path
		configDirectory := instance.Config.GetConfigPath()
		paths := []string{
			configDirectory,
			paths.GetStashHomeDirectory(),
		}
		ffmpegPath, ffprobePath := ffmpeg.GetPaths(paths)

		if ffmpegPath == "" || ffprobePath == "" {
			logger.Infof("couldn't find FFMPEG, attempting to download it")
			if err := ffmpeg.Download(ctx, configDirectory); err != nil {
				msg := `Unable to locate / automatically download FFMPEG

	Check the readme for download links.
	The FFMPEG and FFProbe binaries should be placed in %s

	The error was: %s
	`
				logger.Errorf(msg, configDirectory, err)
				return err
			} else {
				// After download get new paths for ffmpeg and ffprobe
				ffmpegPath, ffprobePath = ffmpeg.GetPaths(paths)
			}
		}

		instance.FFMPEG = ffmpeg.FFMpeg(ffmpegPath)
		instance.FFProbe = ffmpeg.FFProbe(ffprobePath)
	}

	return nil
}

func initLog() *log.Logger {
	config := config.GetInstance()
	l := log.NewLogger()
	l.Init(config.GetLogFile(), config.GetLogOut(), config.GetLogLevel())
	logger.Logger = l

	return l
}

// PostInit initialises the paths, caches and txnManager after the initial
// configuration has been set. Should only be called if the configuration
// is valid.
func (s *Manager) PostInit(ctx context.Context) error {
	if err := s.Config.SetInitialConfig(); err != nil {
		logger.Warnf("could not set initial configuration: %v", err)
	}

	s.Paths = paths.NewPaths(s.Config.GetGeneratedPath())
	s.RefreshConfig()
	s.SessionStore = session.NewStore(s.Config)
	s.PluginCache.RegisterSessionStore(s.SessionStore)

	if err := s.PluginCache.LoadPlugins(); err != nil {
		logger.Errorf("Error reading plugin configs: %s", err.Error())
	}

	s.ScraperCache = instance.initScraperCache()
	writeStashIcon()

	// clear the downloads and tmp directories
	// #1021 - only clear these directories if the generated folder is non-empty
	if s.Config.GetGeneratedPath() != "" {
		const deleteTimeout = 1 * time.Second

		utils.Timeout(func() {
			if err := fsutil.EmptyDir(instance.Paths.Generated.Downloads); err != nil {
				logger.Warnf("could not empty Downloads directory: %v", err)
			}
			if err := fsutil.EmptyDir(instance.Paths.Generated.Tmp); err != nil {
				logger.Warnf("could not empty Tmp directory: %v", err)
			}
		}, deleteTimeout, func(done chan struct{}) {
			logger.Info("Please wait. Deleting temporary files...") // print
			<-done                                                  // and wait for deletion
			logger.Info("Temporary files deleted.")
		})
	}

	if err := database.Initialize(s.Config.GetDatabasePath()); err != nil {
		return err
	}

	if database.Ready() == nil {
		s.PostMigrate(ctx)
	}

	return nil
}

func writeStashIcon() {
	p := FaviconProvider{
		UIBox: ui.UIBox,
	}

	iconPath := filepath.Join(instance.Config.GetConfigPath(), "icon.png")
	err := ioutil.WriteFile(iconPath, p.GetFaviconPng(), 0644)
	if err != nil {
		logger.Errorf("Couldn't write icon file: %s", err.Error())
	}
}

// initScraperCache initializes a new scraper cache and returns it.
func (s *Manager) initScraperCache() *scraper.Cache {
	ret, err := scraper.NewCache(config.GetInstance(), s.TxnManager)

	if err != nil {
		logger.Errorf("Error reading scraper configs: %s", err.Error())
	}

	return ret
}

func (s *Manager) RefreshConfig() {
	s.Paths = paths.NewPaths(s.Config.GetGeneratedPath())
	config := s.Config
	if config.Validate() == nil {
		if err := fsutil.EnsureDir(s.Paths.Generated.Screenshots); err != nil {
			logger.Warnf("could not create directory for Screenshots: %v", err)
		}
		if err := fsutil.EnsureDir(s.Paths.Generated.Vtt); err != nil {
			logger.Warnf("could not create directory for VTT: %v", err)
		}
		if err := fsutil.EnsureDir(s.Paths.Generated.Markers); err != nil {
			logger.Warnf("could not create directory for Markers: %v", err)
		}
		if err := fsutil.EnsureDir(s.Paths.Generated.Transcodes); err != nil {
			logger.Warnf("could not create directory for Transcodes: %v", err)
		}
		if err := fsutil.EnsureDir(s.Paths.Generated.Downloads); err != nil {
			logger.Warnf("could not create directory for Downloads: %v", err)
		}
		if err := fsutil.EnsureDir(s.Paths.Generated.InteractiveHeatmap); err != nil {
			logger.Warnf("could not create directory for Interactive Heatmaps: %v", err)
		}
	}
}

// RefreshScraperCache refreshes the scraper cache. Call this when scraper
// configuration changes.
func (s *Manager) RefreshScraperCache() {
	s.ScraperCache = s.initScraperCache()
}

func setSetupDefaults(input *models.SetupInput) {
	if input.ConfigLocation == "" {
		input.ConfigLocation = filepath.Join(fsutil.GetHomeDirectory(), ".stash", "config.yml")
	}

	configDir := filepath.Dir(input.ConfigLocation)
	if input.GeneratedLocation == "" {
		input.GeneratedLocation = filepath.Join(configDir, "generated")
	}

	if input.DatabaseFile == "" {
		input.DatabaseFile = filepath.Join(configDir, "stash-go.sqlite")
	}
}

func (s *Manager) Setup(ctx context.Context, input models.SetupInput) error {
	setSetupDefaults(&input)
	c := s.Config

	// create the config directory if it does not exist
	// don't do anything if config is already set in the environment
	if !config.FileEnvSet() {
		configDir := filepath.Dir(input.ConfigLocation)
		if exists, _ := fsutil.DirExists(configDir); !exists {
			if err := os.Mkdir(configDir, 0755); err != nil {
				return fmt.Errorf("error creating config directory: %v", err)
			}
		}

		if err := fsutil.Touch(input.ConfigLocation); err != nil {
			return fmt.Errorf("error creating config file: %v", err)
		}

		s.Config.SetConfigFile(input.ConfigLocation)
	}

	// create the generated directory if it does not exist
	if !c.HasOverride(config.Generated) {
		if exists, _ := fsutil.DirExists(input.GeneratedLocation); !exists {
			if err := os.Mkdir(input.GeneratedLocation, 0755); err != nil {
				return fmt.Errorf("error creating generated directory: %v", err)
			}
		}

		s.Config.Set(config.Generated, input.GeneratedLocation)
	}

	// set the configuration
	if !c.HasOverride(config.Database) {
		s.Config.Set(config.Database, input.DatabaseFile)
	}

	s.Config.Set(config.Stash, input.Stashes)
	if err := s.Config.Write(); err != nil {
		return fmt.Errorf("error writing configuration file: %v", err)
	}

	// initialise the database
	if err := s.PostInit(ctx); err != nil {
		return fmt.Errorf("error initializing the database: %v", err)
	}

	s.Config.FinalizeSetup()

	if err := initFFMPEG(ctx); err != nil {
		return fmt.Errorf("error initializing FFMPEG subsystem: %v", err)
	}

	return nil
}

func (s *Manager) validateFFMPEG() error {
	if s.FFMPEG == "" || s.FFProbe == "" {
		return errors.New("missing ffmpeg and/or ffprobe")
	}

	return nil
}

func (s *Manager) Migrate(ctx context.Context, input models.MigrateInput) error {
	// always backup so that we can roll back to the previous version if
	// migration fails
	backupPath := input.BackupPath
	if backupPath == "" {
		backupPath = database.DatabaseBackupPath()
	}

	// perform database backup
	if err := database.Backup(database.DB, backupPath); err != nil {
		return fmt.Errorf("error backing up database: %s", err)
	}

	if err := database.RunMigrations(); err != nil {
		errStr := fmt.Sprintf("error performing migration: %s", err)

		// roll back to the backed up version
		restoreErr := database.RestoreFromBackup(backupPath)
		if restoreErr != nil {
			errStr = fmt.Sprintf("ERROR: unable to restore database from backup after migration failure: %s\n%s", restoreErr.Error(), errStr)
		} else {
			errStr = "An error occurred migrating the database to the latest schema version. The backup database file was automatically renamed to restore the database.\n" + errStr
		}

		return errors.New(errStr)
	}

	// perform post-migration operations
	s.PostMigrate(ctx)

	// if no backup path was provided, then delete the created backup
	if input.BackupPath == "" {
		if err := os.Remove(backupPath); err != nil {
			logger.Warnf("error removing unwanted database backup (%s): %s", backupPath, err.Error())
		}
	}

	return nil
}

func (s *Manager) GetSystemStatus() *models.SystemStatus {
	status := models.SystemStatusEnumOk
	dbSchema := int(database.Version())
	dbPath := database.DatabasePath()
	appSchema := int(database.AppSchemaVersion())
	configFile := s.Config.GetConfigFile()

	if s.Config.IsNewSystem() {
		status = models.SystemStatusEnumSetup
	} else if dbSchema < appSchema {
		status = models.SystemStatusEnumNeedsMigration
	}

	return &models.SystemStatus{
		DatabaseSchema: &dbSchema,
		DatabasePath:   &dbPath,
		AppSchema:      appSchema,
		Status:         status,
		ConfigPath:     &configFile,
	}
}

// Shutdown gracefully stops the manager
func (s *Manager) Shutdown(code int) {
	// stop any profiling at exit
	pprof.StopCPUProfile()

	// TODO: Each part of the manager needs to gracefully stop at some point
	// for now, we just close the database.
	err := database.Close()
	if err != nil {
		logger.Errorf("Error closing database: %s", err)
		if code == 0 {
			os.Exit(1)
		}
	}

	os.Exit(code)
}
