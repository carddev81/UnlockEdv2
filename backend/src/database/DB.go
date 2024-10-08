package database

import (
	"UnlockEdv2/src/models"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	Conn *gorm.DB
}

var TableList = []interface{}{
	&models.User{},
	&models.ProviderPlatform{},
	&models.UserActivity{},
	&models.ProviderUserMapping{},
	&models.LeftMenuLink{},
	&models.Program{},
	&models.Milestone{},
	&models.Outcome{},
	&models.Activity{},
	&models.OidcClient{},
	&models.UserFavorite{},
	&models.Facility{},
	&models.OpenContentProvider{},
	&models.CronJob{},
	&models.RunnableTask{},
}

var validate *validator.Validate

func InitDB(isTesting bool) *DB {
	var gormDb *gorm.DB
	var err error
	if isTesting {
		gormDb, err = gorm.Open(sqlite.Open("file:memdb1?mode=memory&cache=shared"), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to SQLite database:", err)
		}
		log.Println("Connected to the SQLite database in memory")
		MigrateTesting(gormDb)
	} else {
		dsn := os.Getenv("APP_DSN")
		if dsn == "" {
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=allow",
				os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
		}
		gormDb, err = gorm.Open(postgres.New(postgres.Config{
			DSN: dsn,
		}), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
		}
		log.Println("Connected to the PostgreSQL database")
	}
	validate = validator.New(validator.WithRequiredStructEnabled())
	database := &DB{Conn: gormDb}
	SeedDefaultData(gormDb, isTesting)
	if isTesting {
		database.SeedTestData()
	}
	return database
}

func MigrateTesting(db *gorm.DB) {
	for _, table := range TableList {
		log.Printf("Migrating %T table...", table)
		if err := db.AutoMigrate(table); err != nil {
			log.Fatal("Failed to migrate table: ", err)
		}
	}
}

func SeedDefaultData(db *gorm.DB, isTesting bool) {
	var count int64
	if err := db.Model(models.User{}).Where("id = ?", fmt.Sprintf("%d", 1)).Count(&count).Error; err != nil {
		log.Fatal("db transaction failed getting admin user")
	}
	if count == 0 {
		if err := db.Model(models.Facility{}).Where("id = ?", fmt.Sprintf("%d", 1)).Count(&count).Error; err != nil {
			log.Fatal("db transaction failed getting default facility")
		}
		defaultFacility := models.Facility{
			Name: "Default",
		}
		log.Printf("Creating facility: %v", defaultFacility)
		if err := db.Create(&defaultFacility).Error; err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		user := models.User{
			Username:      "SuperAdmin",
			NameFirst:     "Super",
			NameLast:      "Admin",
			Email:         "admin@unlocked.v2",
			PasswordReset: true,
			Role:          "admin",
			Password:      "ChangeMe!",
			FacilityID:    1,
		}
		log.Printf("Creating user: %v", user)
		log.Println("Make sure to sync the Kratos instance if you are freshly migrating")
		err := user.HashPassword()
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}
		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}

		links := []models.LeftMenuLink{}
		if err := json.Unmarshal([]byte(defaultLeftMenuLinks), &links); err != nil {
			log.Fatalf("Failed to unmarshal default left menu links: %v", err)
		}
		if err := db.Create(&links).Error; err != nil {
			log.Fatalf("Failed to create left menu links: %v", err)
		}
	}
}

const defaultLeftMenuLinks = `[{"name":"Unlocked Labs","rank":1,"links":[{"Unlocked Labs Website":"http:\/\/www.unlockedlabs.org\/"},{"Unlocked Labs LinkedIn":"https:\/\/www.linkedin.com\/company\/labs-unlocked\/"}],"created_at":null,"updated_at":null}]`

func (db *DB) SeedTestData() {
	platforms, err := os.ReadFile("test_data/provider_platforms.json")
	if err != nil {
		log.Fatalf("Failed to read test data: %v", err)
	}
	var platform []models.ProviderPlatform
	if err := json.Unmarshal(platforms, &platform); err != nil {
		log.Fatalf("Failed to unmarshal test data: %v", err)
	}
	for _, p := range platform {
		if err := db.Conn.Create(&p).Error; err != nil {
			log.Fatalf("Failed to create platform: %v", err)
		}
	}
	users, err := os.ReadFile("test_data/users.json")
	if err != nil {
		log.Fatalf("Failed to read test data: %v", err)
	}
	var user []models.User
	if err := json.Unmarshal(users, &user); err != nil {
		log.Fatalf("Failed to unmarshal test data: %v", err)
	}
	for idx, u := range user {
		log.Printf("Creating user %s", u.Username)
		if err := db.Conn.Create(&u).Error; err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		for i := 0; i < len(platform); i++ {
			mapping := models.ProviderUserMapping{
				UserID:             u.ID,
				ProviderPlatformID: platform[i].ID,
				ExternalUsername:   u.Username,
				ExternalUserID:     strconv.Itoa(idx),
			}
			if err = db.CreateProviderUserMapping(&mapping); err != nil {
				return
			}
		}
	}
	var programs []models.Program
	progs, err := os.ReadFile("test_data/programs.json")
	if err != nil {
		log.Fatalf("Failed to read test data: %v", err)
	}
	if err := json.Unmarshal(progs, &programs); err != nil {
		log.Fatalf("Failed to unmarshal test data: %v", err)
	}
	for _, p := range programs {
		if err := db.Conn.Create(&p).Error; err != nil {
			log.Fatalf("Failed to create program: %v", err)
		}
	}
	var milestones []models.Milestone
	mstones, err := os.ReadFile("test_data/milestones.json")
	if err != nil {
		log.Fatalf("Failed to read test data: %v", err)
	}
	if err := json.Unmarshal(mstones, &milestones); err != nil {
		log.Fatalf("Failed to unmarshal test data: %v", err)
	}
	for _, m := range milestones {
		if err := db.Conn.Create(&m).Error; err != nil {
			log.Fatalf("Failed to create milestone: %v", err)
		}
	}
	outcomes := []string{"completion", "grade", "certificate", "pathway_completion"}
	for _, user := range user {
		for _, prog := range programs {
			for i := 0; i < 365; i++ {
				if rand.Intn(100)%2 == 0 {
					continue
				}
				startTime := 0
				randTime := rand.Intn(1000)
				// we want activity for the last year
				yearAgo := time.Now().AddDate(-1, 0, 0)
				time := yearAgo.AddDate(0, 0, i)
				activity := models.Activity{
					UserID:     user.ID,
					ProgramID:  prog.ID,
					Type:       "interaction",
					TotalTime:  uint(startTime + randTime),
					TimeDelta:  uint(randTime),
					ExternalID: strconv.Itoa(rand.Intn(1000)),
					CreatedAt:  time,
				}
				startTime += randTime
				if err := db.Conn.Create(&activity).Error; err != nil {
					log.Fatalf("Failed to create activity: %v", err)
				}
			}
			outcome := models.Outcome{
				ProgramID: prog.ID,
				Type:      models.OutcomeType(outcomes[rand.Intn(len(outcomes))]),
			}
			if err := db.Conn.Create(&outcome).Error; err != nil {
				log.Fatalf("Failed to create outcome: %v", err)
			}
		}
	}
}
