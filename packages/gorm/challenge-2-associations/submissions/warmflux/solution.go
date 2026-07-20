package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User represents a user in the blog system
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Posts     []Post `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Post represents a blog post
type Post struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID"`
	Tags      []Tag  `gorm:"many2many:post_tags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tag represents a tag for categorizing posts
type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique;not null"`
	Posts []Post `gorm:"many2many:post_tags;"`
}

// ConnectDB establishes a connection to the SQLite database and auto-migrates the models
func ConnectDB() (*gorm.DB, error) {
	// TODO: Implement database connection with auto-migration
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	models := []interface{}{
		&User{},
		&Post{},
		&Tag{},
	}
	err = db.AutoMigrate(models...)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CreateUserWithPosts creates a new user with associated posts
func CreateUserWithPosts(db *gorm.DB, user *User) error {
	// TODO: Implement user creation with posts
	return db.Create(&user).Error
}

// GetUserWithPosts retrieves a user with all their posts preloaded
func GetUserWithPosts(db *gorm.DB, userID uint) (*User, error) {
	// TODO: Implement user retrieval with posts
	var user User
	result := db.Preload("Posts").First(&user, userID)
	return &user, result.Error
}

// CreatePostWithTags creates a new post with specified tags
func CreatePostWithTags(db *gorm.DB, post *Post, tagNames []string) error {
	// TODO: Implement post creation with tags
	var tags []Tag
	for _, name := range tagNames {
		tags = append(tags, Tag{Name: name})
	}
	err := db.Create(&post)
	if err.Error != nil {
		return err.Error
	}
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	tx.Model(&post).Association("Tags").Append(&tags)
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

// GetPostsByTag retrieves all posts that have a specific tag
func GetPostsByTag(db *gorm.DB, tagName string) ([]Post, error) {
	// TODO: Implement posts retrieval by tag
	var postsFound []Post

	err := db.Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Joins("JOIN tags ON post_tags.tag_id = tags.id").
		Where("tags.name = ?", tagName).
		Find(&postsFound).Error
	if err != nil {
		return nil, err
	}
	return postsFound, nil
}

// AddTagsToPost adds tags to an existing post
func AddTagsToPost(db *gorm.DB, postID uint, tagNames []string) error {
	// TODO: Implement adding tags to existing post
	var tags []Tag
	for _, name := range tagNames {
		tags = append(tags, Tag{Name: name})
	}
	var post Post
	result := db.First(&post, postID)
	if result.Error != nil {
		return result.Error
	}
	db.Model(&post).Association("Tags").Append(&tags)
	return nil
}

// GetPostWithUserAndTags retrieves a post with user and tags preloaded
func GetPostWithUserAndTags(db *gorm.DB, postID uint) (*Post, error) {
	// TODO: Implement post retrieval with user and tags
	var post Post
	tx := db.Preload("User").Preload("Tags").First(&post, postID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &post, nil
}