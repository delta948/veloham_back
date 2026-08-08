package models

import (
	"strings"
	"time"
)

type User struct {
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Username      string    `json:"username"`
	Email         string    `gorm:"uniqueIndex" json:"-"`
	Phone         string    `gorm:"uniqueIndex" json:"-"`
	PasswordHash  string    `json:"-"`
	AvatarURL     string    `json:"avatar_url"`
	City          string    `json:"city"`
	Contact       string    `json:"contact"`
	Role          string    `json:"role"`
	IsBlocked     bool      `json:"is_blocked"`
	BlockedReason string    `json:"blocked_reason,omitempty"`
	Rating        float64   `json:"rating"`
	CreatedAt     time.Time `json:"created_at"`
}

type PendingRegistration struct {
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Username      string    `json:"username"`
	Email         string    `gorm:"uniqueIndex" json:"email"`
	Phone         *string   `gorm:"uniqueIndex" json:"phone,omitempty"`
	City          string    `json:"city"`
	Contact       string    `json:"contact"`
	PasswordHash  string    `json:"-"`
	ProviderToken string    `json:"-"`
	CodeHash      string    `json:"-"`
	Attempts      int       `json:"-"`
	ResendCount   int       `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	ResendAfter   time.Time `json:"resend_after"`
}

type PasswordReset struct {
	ID          string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID      string    `gorm:"type:uuid;index" json:"-"`
	CodeHash    string    `json:"-"`
	Attempts    int       `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ResendAfter time.Time `json:"resend_after"`
}

type UserBlockEvent struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;index" json:"user_id"`
	AdminID   string    `gorm:"type:uuid;index" json:"admin_id"`
	Action    string    `json:"action"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
	Admin     User      `gorm:"foreignKey:AdminID" json:"admin"`
}

type PrivateUser struct {
	User
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

func UserWithEmail(user User) PrivateUser {
	email := user.Email
	if strings.HasSuffix(email, "@phone.veloham.local") {
		email = ""
	}
	return PrivateUser{User: user, Email: email, Phone: user.Phone}
}

type Listing struct {
	ID                    string          `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID                string          `gorm:"type:uuid" json:"user_id"`
	User                  User            `json:"user"`
	Title                 string          `json:"title"`
	Description           string          `json:"description"`
	Price                 int             `json:"price"`
	InitialPrice          int             `json:"initial_price"`
	PreviousPrice         *int            `gorm:"->;-:migration" json:"previous_price,omitempty"`
	LastPriceChangeAt     *time.Time      `gorm:"->;-:migration" json:"last_price_change_at,omitempty"`
	City                  string          `json:"city"`
	Brand                 string          `json:"brand"`
	Category              string          `json:"category"`
	BikeType              string          `json:"bike_type"`
	Condition             string          `json:"condition"`
	FrameSize             string          `json:"frame_size"`
	FrameSizeText         string          `json:"frame_size_text"`
	RiderMin              int             `gorm:"column:rider_height_min" json:"rider_height_min"`
	RiderMax              int             `gorm:"column:rider_height_max" json:"rider_height_max"`
	RecommendedHeightMin  int             `json:"recommended_height_min"`
	RecommendedHeightMax  int             `json:"recommended_height_max"`
	DealType              string          `json:"deal_type"`
	Labels                []string        `gorm:"type:jsonb;serializer:json" json:"labels"`
	IsUrgent              bool            `json:"is_urgent"`
	IsBargain             bool            `json:"is_bargain"`
	IsExchange            bool            `json:"is_exchange"`
	ExtraPaymentFromMe    bool            `json:"extra_payment_from_me"`
	ExtraPaymentFromBuyer bool            `json:"extra_payment_from_buyer"`
	Status                string          `json:"status"`
	Views                 int             `json:"views"`
	Images                []ListingImage  `json:"images"`
	BuildCard             BuildCard       `json:"build_card"`
	MatchPref             MatchPreference `json:"match_preference"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type ListingPriceHistory struct {
	ID               string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ListingID        string    `gorm:"type:uuid;index" json:"listing_id"`
	OldPrice         int       `json:"old_price"`
	NewPrice         int       `json:"new_price"`
	ChangedAt        time.Time `json:"changed_at"`
	ChangedBy        string    `gorm:"type:uuid" json:"changed_by"`
	IPAddress        string    `json:"ip_address"`
	Suspicious       bool      `json:"suspicious"`
	SuspiciousReason string    `json:"suspicious_reason,omitempty"`
	Listing          Listing   `json:"listing,omitempty"`
	User             User      `gorm:"foreignKey:ChangedBy" json:"changed_by_user,omitempty"`
}

func (ListingPriceHistory) TableName() string { return "listing_price_history" }

type Notification struct {
	ID             string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID         string    `gorm:"type:uuid" json:"user_id"`
	ListingID      string    `gorm:"type:uuid" json:"listing_id"`
	PriceHistoryID string    `gorm:"type:uuid" json:"price_history_id"`
	Type           string    `json:"type"`
	Message        string    `json:"message"`
	Link           string    `json:"link"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

type ListingPlacement struct {
	ID                string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID            string     `gorm:"type:uuid;index" json:"user_id"`
	ListingID         *string    `gorm:"type:uuid;uniqueIndex" json:"listing_id,omitempty"`
	Kind              string     `json:"kind"`
	TargetStatus      string     `json:"target_status"`
	Amount            int        `json:"amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	Provider          string     `json:"provider,omitempty"`
	ProviderPaymentID string     `json:"provider_payment_id,omitempty"`
	CheckoutURL       string     `json:"checkout_url,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
}

type ListingImage struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ListingID string `gorm:"type:uuid" json:"listing_id"`
	ImageURL  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
}

type Favorite struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;uniqueIndex:idx_favorite_user_listing" json:"user_id"`
	ListingID string    `gorm:"type:uuid;uniqueIndex:idx_favorite_user_listing" json:"listing_id"`
	Listing   Listing   `json:"listing"`
	CreatedAt time.Time `json:"created_at"`
}

type Chat struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	BuyerID   string    `gorm:"type:uuid;uniqueIndex:idx_chat_unique" json:"buyer_id"`
	SellerID  string    `gorm:"type:uuid;uniqueIndex:idx_chat_unique" json:"seller_id"`
	ListingID string    `gorm:"type:uuid;uniqueIndex:idx_chat_unique" json:"listing_id"`
	Buyer     User      `json:"buyer"`
	Seller    User      `json:"seller"`
	Listing   Listing   `json:"listing"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ChatID    string    `gorm:"type:uuid" json:"chat_id"`
	SenderID  string    `gorm:"type:uuid" json:"sender_id"`
	Sender    User      `json:"sender"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	IsRead    bool      `json:"is_read"`
}

type BuildCard struct {
	ID             string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ListingID      string `gorm:"type:uuid;uniqueIndex" json:"listing_id"`
	Frame          string `json:"frame"`
	Size           string `json:"size"`
	Fork           string `json:"fork"`
	Wheels         string `json:"wheels"`
	Hubs           string `json:"hubs"`
	Tires          string `json:"tires"`
	Handlebar      string `json:"handlebar"`
	Stem           string `json:"stem"`
	Saddle         string `json:"saddle"`
	Cranks         string `json:"cranks"`
	BottomBracket  string `json:"bottom_bracket"`
	Chain          string `json:"chain"`
	Cog            string `json:"cog"`
	Brakes         string `json:"brakes"`
	Weight         string `json:"weight"`
	FrameCondition string `json:"frame_condition"`
	Defects        string `json:"defects"`
	Documents      bool   `json:"documents"`
}

type MatchPreference struct {
	ID              string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ListingID       string `gorm:"type:uuid;uniqueIndex" json:"listing_id"`
	ExchangeEnabled bool   `json:"exchange_enabled"`
	Wants           string `json:"wants"`
	Categories      string `json:"categories"`
	MinPrice        int    `json:"min_price"`
	MaxPrice        int    `json:"max_price"`
	CanAddCash      bool   `json:"can_add_cash"`
	MaxCashAdd      int    `json:"max_cash_add"`
	SameCityOnly    bool   `json:"same_city_only"`
}

type Review struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	SellerID  string    `gorm:"type:uuid" json:"seller_id"`
	AuthorID  string    `gorm:"type:uuid" json:"author_id"`
	ListingID string    `gorm:"type:uuid" json:"listing_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	Author    User      `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type Report struct {
	ID         string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ReporterID string    `gorm:"type:uuid" json:"reporter_id"`
	ListingID  string    `gorm:"type:uuid" json:"listing_id"`
	SellerID   string    `gorm:"type:uuid" json:"seller_id"`
	Reason     string    `json:"reason"`
	Text       string    `json:"text"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Reporter   User      `json:"reporter"`
	Listing    Listing   `json:"listing"`
}

type WantedRequest struct {
	ID                string        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID            string        `gorm:"type:uuid" json:"user_id"`
	User              User          `json:"user"`
	Title             string        `json:"title"`
	Category          string        `json:"category"`
	MinBudget         int           `json:"min_budget"`
	MaxBudget         int           `json:"max_budget"`
	BudgetMin         int           `json:"budget_min"`
	BudgetMax         int           `json:"budget_max"`
	City              string        `json:"city"`
	FrameSize         string        `json:"frame_size"`
	RiderHeight       int           `json:"rider_height"`
	Height            int           `json:"height"`
	PreferredBikeType string        `json:"preferred_bike_type"`
	Description       string        `json:"description"`
	Status            string        `json:"status"`
	Offers            []WantedOffer `gorm:"foreignKey:WantedID" json:"offers"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type WantedOffer struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	WantedID  string    `gorm:"type:uuid" json:"wanted_id"`
	SellerID  string    `gorm:"type:uuid" json:"seller_id"`
	ListingID string    `gorm:"type:uuid" json:"listing_id"`
	Message   string    `json:"message"`
	Seller    User      `json:"seller"`
	Listing   Listing   `json:"listing"`
	CreatedAt time.Time `json:"created_at"`
}
