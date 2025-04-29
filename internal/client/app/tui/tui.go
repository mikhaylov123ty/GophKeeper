package tui

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type Model struct {
	currentScreen Screen
}

type Screen interface {
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
}

type Category string

const (
	TextCategory  Category = "Text"
	CredsCategory Category = "Creds"
	FileCategory  Category = "Files"
	CardCategory  Category = "Cards"
	ExitCategory  Category = "Exit" // New exit category

	listHeight = 15

	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
)

type MainMenu struct {
	categories []Category
	cursor     int
	manager    *ItemManager
	nextScreen Screen
}

type ActionsMenu struct {
	options     []string
	cursor      int
	category    Category
	itemManager *ItemManager
	backScreen  Screen
}

type ErrorScreen struct {
	backScreen Screen
	err        error
}

type ItemManager struct {
	metaItems  map[Category][]*MetaItem
	grpcClient *grpc.Client
	userID     string
}

func NewItemManager(grpcClient *grpc.Client) (*Model, error) {
	im := ItemManager{
		metaItems:  map[Category][]*MetaItem{},
		grpcClient: grpcClient,
	}

	mainMenu := MainMenu{
		categories: []Category{TextCategory, CredsCategory, FileCategory, CardCategory, ExitCategory}, // Include exit category
		cursor:     0,
		manager:    &im,
	}

	auth := NewAuthScreen(&mainMenu, &im)

	return auth, nil
}

func (im *ItemManager) postItemData(data []byte, dataID string, metaData *pb.MetaData) (*pb.PostItemDataResponse, error) {
	encryptedData, err := encryptData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	resp, err := im.grpcClient.Handlers.ItemDataHandler.PostItemData(context.Background(),
		&pb.PostItemDataRequest{
			Data:     encryptedData,
			DataId:   dataID,
			MetaData: metaData,
		})
	if err != nil {
		return nil, fmt.Errorf("post item failed:  %w,", err)
	}

	return resp, err
}

func (im *ItemManager) getItemData(dataID string) (string, error) {
	response, err := im.grpcClient.Handlers.ItemDataHandler.GetItemData(context.Background(), &pb.GetItemDataRequest{
		DataId: dataID,
	})
	if err != nil {
		return "", fmt.Errorf("could not get text data: %w", err)
	}

	decryptedData, err := deryptData(response.Data)
	if err != nil {
		return "", fmt.Errorf("failed decrypt data: %w", err)
	}

	return string(decryptedData), nil

}

func (im *ItemManager) syncMeta() error {
	metaItems, err := im.grpcClient.Handlers.MetaDataHandler.GetMetaData(context.Background(),
		&pb.GetMetaDataRequest{UserId: im.userID})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				fmt.Println(`NOT FOUND`, e.Message())
				return nil
			} else {
				fmt.Println(e.Code(), e.Message())
			}
		} else {
			fmt.Printf("Не получилось распарсить ошибку %v", err)
			return err
		}
	}

	for _, metaItem := range metaItems.Items {
		id, err := uuid.Parse(metaItem.GetId())
		if err != nil {
			return fmt.Errorf("invalid meta item id: %s", metaItem.GetId())
		}
		im.metaItems[Category(metaItem.DataType)] = append(im.metaItems[Category(metaItem.DataType)], &MetaItem{
			Id:          id,
			Title:       metaItem.GetTitle(),
			Description: metaItem.GetDescription(),
			dataID:      metaItem.GetDataId(),
			Created:     metaItem.GetCreated(),
			Modified:    metaItem.GetModified(),
		})
	}

	return nil
}

func (im *ItemManager) deleteItem(metaItemID uuid.UUID, category Category, dataID string) error {
	resp, err := im.grpcClient.Handlers.MetaDataHandler.DeleteMetaData(context.Background(), &pb.DeleteMetaDataRequest{
		MetadataId:   metaItemID.String(),
		MetadataType: string(category),
		DataId:       dataID,
	})
	if err != nil && resp.GetError() != "" {
		return fmt.Errorf("could not delete meta data: %w", err)
	}

	for i, v := range im.metaItems[category] {
		if v.Id == metaItemID {
			im.metaItems[category] = append(im.metaItems[category][:i], im.metaItems[category][i+1:]...)
		}
	}

	return nil
}

type DeleteItemScreen struct {
	itemManager *ItemManager
	category    Category
	deleteIndex int
	backScreen  Screen
}

const (
	ViewOption = "View items"
	AddOption  = "Add an item"
	BackOption = "Back"
)

// Style definitions
var (
	cursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))                     // Bright purple
	selectedStyle   = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("2")).Bold(true) // Green
	unselectedStyle = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("7"))            // White
	backgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("245"))                     // Grey background
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("65"))           // Bold yellow
	separatorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))                     // Light grey
)

func (m MainMenu) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.cursor = (m.cursor + 1) % len(m.categories)
		case "up":
			m.cursor = (m.cursor - 1 + len(m.categories)) % len(m.categories)
		case "enter":
			category := m.categories[m.cursor]
			if category == ExitCategory {
				return m, tea.Quit // Exit the application
			}
			m.nextScreen = &ActionsMenu{
				options:     []string{ViewOption, AddOption, BackOption},
				category:    category,
				itemManager: m.manager,
				backScreen:  m,
			}
			return m.nextScreen, nil
		}
	}

	return m, nil
}

func (m MainMenu) View() string {
	s := titleStyle.Render("Main Menu:\n\n")
	for i, category := range m.categories {
		if m.cursor == i {
			s += selectedStyle.Render(fmt.Sprintf("[x] %s\n", string(category))) // Selected option with color
		} else {
			s += unselectedStyle.Render(fmt.Sprintf("[ ] %s\n", category)) // Unselected option with color
		}
	}
	s += navigateFooter()
	return s
}

func (cm ActionsMenu) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			cm.cursor = (cm.cursor + 1) % len(cm.options)
		case "up":
			cm.cursor = (cm.cursor - 1 + len(cm.options)) % len(cm.options)
		case "enter":
			switch cm.cursor {
			case 0: // View items
				switch cm.category {
				case TextCategory:
					//TODO make screens construct to reuse
					return &ViewTextItemsScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case CredsCategory:
					return &ViewCredsItemsScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case FileCategory:
					return &ViewBinaryItemsScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case CardCategory:
					return &ViewBankCardItemsScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				}
			case 1: // Add item
				switch cm.category {
				case TextCategory:
					return &AddTextItemScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case CredsCategory:
					return &AddCredsItemScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case FileCategory:
					return &AddBinaryItemScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case CardCategory:
					return &AddBankCardItemScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				}
			case 2: // Back
				return cm.backScreen, nil
			}
		case "esc": // Go back to main menu
			return cm.backScreen, nil
		}
	}
	return cm, nil
}

func (cm ActionsMenu) View() string {
	s := titleStyle.Render("Category Menu:\n\n")
	for i, option := range cm.options {
		if cm.cursor == i {
			s += selectedStyle.Render(fmt.Sprintf("[x] %s\n", option)) // Selected option with color
		} else {
			s += unselectedStyle.Render(fmt.Sprintf("[ ] %s\n", option)) // Unselected option with color
		}
	}

	s += navigateFooter()
	return s
}

func (screen *ErrorScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		default:
			return screen.backScreen, nil
		}
	}
	return screen.backScreen, nil
}

func (screen *ErrorScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sError%s\n"+
			"=======================%s"+
			"%s%s%s\n",
		ColorBold, ColorReset,
		separator,
		ColorGreen, screen.err.Error(), ColorReset)
}

func (m Model) Init() tea.Cmd {
	return nil // No command to run, initial screen is set.
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	nextScreen, cmd := m.currentScreen.Update(msg)
	m.currentScreen = nextScreen
	return m, cmd
}

func (m Model) View() string {
	return m.currentScreen.View()
}

func encryptData(data []byte) ([]byte, error) {
	// Пропуск обработки, если флаг не задан
	if config.GetKeys().PublicCert == "" {
		return data, nil
	}

	var err error
	var publicPEM []byte

	publicPEM, err = os.ReadFile(config.GetKeys().PublicCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read public certificate: %w", err)
	}

	publicCertBlock, _ := pem.Decode(publicPEM)
	certHash := []byte(createHash(publicCertBlock.Bytes))

	block, err := aes.NewCipher(certHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func deryptData(body []byte) ([]byte, error) {
	if config.GetKeys().PublicCert == "" {
		return body, nil
	}

	var err error
	var publicPEM []byte

	publicPEM, err = os.ReadFile(config.GetKeys().PublicCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read public certificate: %w", err)
	}

	publicCertBlock, _ := pem.Decode(publicPEM)
	certHash := []byte(createHash(publicCertBlock.Bytes))

	block, err := aes.NewCipher(certHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	decodedBody, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to decode body: %w", err)
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := decodedBody[:nonceSize], decodedBody[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

func createHash(key []byte) string {
	hasher := sha256.New()
	hasher.Write(key)
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))[:32]
}

func navigateFooter() string {
	return backgroundStyle.Render(separatorStyle.Render("Use arrow keys to navigate and enter to select.\n")) // Navigation instructions
}

func addItemsFooter() string {
	return backgroundStyle.Render(separatorStyle.Render("Press Enter to save, CTRL+Q to cancel, or Backspace to delete the last character.\n"))
}

func listItemsFooter() string {
	return backgroundStyle.Render(separatorStyle.Render("\nUse arrow keys to navigate. E to edit. D to delete. Enter to select. CTRL+Q to cancel.\n"))
}

func itemDataFooter() string {
	return backgroundStyle.Render(separatorStyle.Render("\nPress CTRL+Q to return.\n"))
}
