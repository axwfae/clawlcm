package store

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/axwfae/clawlcm/db"
	"github.com/axwfae/clawlcm/types"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetDB() *gorm.DB {
	return s.db
}

func (s *Store) CreateConversation(sessionKey, sessionID string) (int64, error) {
	conv := db.ConversationModel{
		SessionKey:   sessionKey,
		SessionID:    sessionID,
		MessageCount: 0,
		TokenCount:   0,
	}
	err := s.db.Create(&conv).Error
	return conv.ID, err
}

func (s *Store) GetConversationBySessionKey(sessionKey string) (*types.Conversation, error) {
	var model db.ConversationModel
	err := s.db.Where("session_key = ?", sessionKey).First(&model).Error
	if err != nil {
		return nil, err
	}
	return toConversation(&model), nil
}

func (s *Store) GetConversationByID(id int64) (*types.Conversation, error) {
	var model db.ConversationModel
	err := s.db.First(&model, id).Error
	if err != nil {
		return nil, err
	}
	return toConversation(&model), nil
}

func (s *Store) GetAllConversations() ([]types.Conversation, error) {
	var models []db.ConversationModel
	err := s.db.Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	convs := make([]types.Conversation, len(models))
	for i, m := range models {
		convs[i] = *toConversation(&m)
	}
	return convs, nil
}

func (s *Store) GetMessageByID(id int64) (*types.MessageRecord, error) {
	var model db.MessageModel
	err := s.db.First(&model, id).Error
	if err != nil {
		return nil, err
	}
	records := toMessages([]db.MessageModel{model})
	return &records[0], nil
}

func (s *Store) UpdateConversationStats(id int64, messageCount, tokenCount int) error {
	return s.db.Model(&db.ConversationModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"message_count": messageCount,
		"token_count":   tokenCount,
	}).Error
}

func (s *Store) CreateMessage(conversationID int64, ordinal int, role, content string, tokenCount int) (int64, error) {
	msg := db.MessageModel{
		ConversationID: conversationID,
		Ordinal:        ordinal,
		Role:           role,
		Content:        content,
		TokenCount:     tokenCount,
	}
	err := s.db.Create(&msg).Error
	return msg.ID, err
}

func (s *Store) GetMessages(conversationID int64, limit, offset int) ([]types.MessageRecord, error) {
	var models []db.MessageModel
	err := s.db.Where("conversation_id = ?", conversationID).Order("ordinal ASC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toMessages(models), nil
}

func (s *Store) GetMessageCount(conversationID int64) (int, error) {
	var count int64
	err := s.db.Model(&db.MessageModel{}).Where("conversation_id = ?", conversationID).Count(&count).Error
	return int(count), err
}

func (s *Store) DeleteMessage(id int64) error {
	return s.db.Delete(&db.MessageModel{}, id).Error
}

func (s *Store) DeleteSummary(id int64) error {
	return s.db.Delete(&db.SummaryModel{}, id).Error
}

func (s *Store) UpdateMessageOrdinal(id int64, ordinal int) error {
	return s.db.Model(&db.MessageModel{}).Where("id = ?", id).Update("ordinal", ordinal).Error
}

func (s *Store) GetTotalTokens(conversationID int64) (int, error) {
	var total int64
	err := s.db.Model(&db.MessageModel{}).Where("conversation_id = ?", conversationID).Select("COALESCE(SUM(token_count), 0)").Scan(&total).Error
	return int(total), err
}

func (s *Store) CreateSummary(conversationID int64, summaryType types.SummaryType, depth int, content string, tokenCount, sourceTokens, ordinal int, parentIDs, sourceIDs []int64) (int64, error) {
	parentJSON, _ := json.Marshal(parentIDs)
	sourceJSON, _ := json.Marshal(sourceIDs)
	summary := db.SummaryModel{
		ConversationID: conversationID,
		SummaryType:    string(summaryType),
		Depth:          depth,
		Content:        content,
		TokenCount:     tokenCount,
		SourceTokens:   sourceTokens,
		Ordinal:        ordinal,
		ParentIDs:      string(parentJSON),
		SourceIDs:      string(sourceJSON),
	}
	err := s.db.Create(&summary).Error
	return summary.ID, err
}

func (s *Store) GetSummaries(conversationID int64) ([]types.SummaryRecord, error) {
	var models []db.SummaryModel
	err := s.db.Where("conversation_id = ?", conversationID).Order("ordinal ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toSummaries(models), nil
}

func (s *Store) GetSummaryByID(id int64) (*types.SummaryRecord, error) {
	var model db.SummaryModel
	err := s.db.First(&model, id).Error
	if err != nil {
		return nil, err
	}
	records := toSummaries([]db.SummaryModel{model})
	if len(records) == 0 {
		return nil, fmt.Errorf("summary not found")
	}
	return &records[0], nil
}

func (s *Store) GetLeafSummaries(conversationID int64) ([]types.SummaryRecord, error) {
	var models []db.SummaryModel
	err := s.db.Where("conversation_id = ? AND summary_type = ?", conversationID, "leaf").Order("ordinal ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toSummaries(models), nil
}

func (s *Store) GetCondensedSummaries(conversationID int64) ([]types.SummaryRecord, error) {
	var models []db.SummaryModel
	err := s.db.Where("conversation_id = ? AND summary_type = ?", conversationID, "condensed").Order("ordinal ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toSummaries(models), nil
}

func (s *Store) GetSummariesByDepth(conversationID int64, depth int) ([]types.SummaryRecord, error) {
	var models []db.SummaryModel
	err := s.db.Where("conversation_id = ? AND depth = ?", conversationID, depth).Order("ordinal ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toSummaries(models), nil
}

func (s *Store) CreateContextItem(conversationID int64, itemType types.ContextItemType, itemID int64, ordinal, tokenCount int, keywords string) (int64, error) {
	item := db.ContextItemModel{
		ConversationID: conversationID,
		ItemType:       string(itemType),
		ItemID:         itemID,
		Ordinal:        ordinal,
		TokenCount:     tokenCount,
		Keywords:       keywords,
	}
	err := s.db.Create(&item).Error
	return item.ID, err
}

func (s *Store) GetContextItems(conversationID int64) ([]types.ContextItemRecord, error) {
	var models []db.ContextItemModel
	err := s.db.Where("conversation_id = ?", conversationID).Order("ordinal ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toContextItems(models), nil
}

func (s *Store) ClearContextItems(conversationID int64) error {
	return s.db.Where("conversation_id = ?", conversationID).Delete(&db.ContextItemModel{}).Error
}

func toConversation(m *db.ConversationModel) *types.Conversation {
	return &types.Conversation{
		ID:           m.ID,
		SessionKey:   m.SessionKey,
		SessionID:    m.SessionID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		MessageCount: m.MessageCount,
		TokenCount:   m.TokenCount,
	}
}

func toMessages(models []db.MessageModel) []types.MessageRecord {
	records := make([]types.MessageRecord, len(models))
	for i, m := range models {
		records[i] = types.MessageRecord{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Ordinal:        m.Ordinal,
			Role:           m.Role,
			Content:        m.Content,
			TokenCount:     m.TokenCount,
			CreatedAt:      m.CreatedAt,
		}
	}
	return records
}

func toSummaries(models []db.SummaryModel) []types.SummaryRecord {
	records := make([]types.SummaryRecord, len(models))
	for i, m := range models {
		var parentIDs, sourceIDs []int64
		json.Unmarshal([]byte(m.ParentIDs), &parentIDs)
		json.Unmarshal([]byte(m.SourceIDs), &sourceIDs)
		records[i] = types.SummaryRecord{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			SummaryType:    types.SummaryType(m.SummaryType),
			Depth:          m.Depth,
			Content:        m.Content,
			TokenCount:     m.TokenCount,
			SourceTokens:   m.SourceTokens,
			Ordinal:        m.Ordinal,
			ParentIDs:      parentIDs,
			SourceIDs:      sourceIDs,
			CreatedAt:      m.CreatedAt,
		}
	}
	return records
}

func toContextItems(models []db.ContextItemModel) []types.ContextItemRecord {
	records := make([]types.ContextItemRecord, len(models))
	for i, m := range models {
		records[i] = types.ContextItemRecord{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			ItemType:       types.ContextItemType(m.ItemType),
			ItemID:         m.ItemID,
			Ordinal:        m.Ordinal,
			TokenCount:     m.TokenCount,
			Keywords:       m.Keywords,
			CreatedAt:      m.CreatedAt,
		}
	}
	return records
}

type SearchResult struct {
	Item  types.ContextItemRecord
	Score float64
}
