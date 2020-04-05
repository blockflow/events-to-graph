package generators

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/elaletovic/events-to-graph/models"

	"github.com/ThreeDotsLabs/watermill"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/brianvoe/gofakeit"
)

var (
	initialEvents            = []string{models.ItemViewed, models.UserAddressValidated, models.UserAddressValidationFailed}
	itemViewedAfterEvents    = []string{models.ItemPurchased, models.ItemDropped, models.Nothing}
	itemPurchasedAfterEvents = []string{models.ItemDelivered, models.ItemNotDelivered}
	// CreateTopic --
	CreateTopic = "create_topic"
	// InitialEventsTopic --
	InitialEventsTopic = "initial_events_topic"
	// CheckoutTopic --
	CheckoutTopic = "checkout_topic"
	// DeliveryTopic --
	DeliveryTopic = "delivery_topic"
)

// GenerateEvents --
func GenerateEvents(publisher message.Publisher) {
	//first generate some users and items
	users := generateUsers(30, publisher)
	items := generateItems(20, publisher)

	for {
		eventType := gofakeit.RandString(initialEvents)
		var eventObj interface{}
		switch eventType {
		case models.ItemViewed:
			eventObj = models.ItemViewedPayload{
				ItemID: items[gofakeit.Number(0, len(items)-1)].ID,
				UserID: users[gofakeit.Number(0, len(users)-1)].ID,
			}
		default:
			continue
		}

		eventPayload, err := json.Marshal(&eventObj)
		if err != nil {
			log.Printf("generateEvents: error while marshalling event payload, error %v\n", err)
			continue
		}

		publish(eventType, InitialEventsTopic, "GenerateEvents", eventPayload, publisher)

		time.Sleep(500 * time.Millisecond)
	}
}

// GeneratorHandler --
type GeneratorHandler struct {
}

// InitialEventsHandler --
func (gh GeneratorHandler) InitialEventsHandler(msg *message.Message) ([]*message.Message, error) {
	event := models.Event{}
	err := json.Unmarshal(msg.Payload, &event)
	if err != nil {
		log.Printf("failed to unmarshal initial events. Error %v, payload: %v\n", err, string(msg.Payload))
		return nil, err
	}
	newEvent := models.Event{
		CreatedAt: time.Now().Unix(),
		Type:      gofakeit.RandString(itemViewedAfterEvents),
	}
	switch event.Type {
	case models.ItemViewed:
		eventPayload := models.ItemViewedPayload{}
		err = json.Unmarshal(event.Payload, &eventPayload)
		if err != nil {
			log.Printf("failed to unmarshal event payload. Error %v, payload: %v\n", err, string(event.Payload))
			return nil, err
		}

		var newEventObj interface{}
		switch newEvent.Type {
		case models.ItemPurchased:
			newEventObj = models.ItemPurchasedPayload{
				ItemID:   eventPayload.ItemID,
				UserID:   eventPayload.UserID,
				Quantity: gofakeit.Number(1, 5),
			}
		case models.ItemDropped:
			newEventObj = models.ItemDroppedPayload{
				ItemID:   eventPayload.ItemID,
				UserID:   eventPayload.UserID,
				Quantity: gofakeit.Number(1, 5),
			}
		case models.Nothing:
			log.Println("returning Nothing")
			return nil, nil
		}

		newEventPayload, err := json.Marshal(&newEventObj)
		if err != nil {
			log.Printf("failed to marshal new event payload. Error %v\n", err)
			return nil, err
		}
		newEvent.Payload = newEventPayload
		payload, err := json.Marshal(&newEvent)
		if err != nil {
			log.Printf("error while marshalling main payload, error %v\n", err)
			return nil, err
		}

		newMsg := message.NewMessage(watermill.NewUUID(), payload)
		return message.Messages{newMsg}, nil
	}
	return nil, nil
}

// PurchasedEventsHandler --
func (gh GeneratorHandler) PurchasedEventsHandler(msg *message.Message) ([]*message.Message, error) {
	event := models.Event{}
	err := json.Unmarshal(msg.Payload, &event)
	if err != nil {
		log.Printf("PurchasedEventsHandler: failed to unmarshal initial events. Error %v, payload: %v\n", err, string(msg.Payload))
		return nil, err
	}
	newEvent := models.Event{
		CreatedAt: time.Now().Unix(),
		Type:      gofakeit.RandString(itemViewedAfterEvents),
	}
	switch event.Type {
	case models.ItemPurchased:
		eventPayload := models.ItemPurchasedPayload{}
		err = json.Unmarshal(event.Payload, &eventPayload)
		if err != nil {
			log.Printf("PurchasedEventsHandler: failed to unmarshal event payload. Error %v, payload: %v\n", err, string(event.Payload))
			return nil, err
		}

		var newEventObj interface{}
		switch newEvent.Type {
		case models.ItemDelivered:
			newEventObj = models.ItemDeliveredPayload{
				Address: fmt.Sprintf("%s, %s, %s", gofakeit.Street(), gofakeit.City(), gofakeit.Country()),
				UserID:  eventPayload.UserID,
				ItemID:  eventPayload.ItemID,
			}
		case models.ItemNotDelivered:
			newEventObj = models.ItemNotDeliveredPayload{
				ItemID:  eventPayload.ItemID,
				UserID:  eventPayload.UserID,
				Address: fmt.Sprintf("%s, %s, %s", gofakeit.Street(), gofakeit.City(), gofakeit.Country()),
				Reason:  gofakeit.RandString([]string{"fake", "not occupied by user"}),
			}
		}

		newEventPayload, err := json.Marshal(&newEventObj)
		if err != nil {
			log.Printf("PurchasedEventsHandler: failed to marshal new event payload. Error %v\n", err)
			return nil, err
		}
		newEvent.Payload = newEventPayload
		payload, err := json.Marshal(&newEvent)
		if err != nil {
			log.Printf("PurchasedEventsHandler: error while marshalling main payload, error %v\n", err)
			return nil, err
		}

		newMsg := message.NewMessage(watermill.NewUUID(), payload)
		return message.Messages{newMsg}, nil
	}
	return nil, nil
}

func generateUsers(numberOfUsers int, publisher message.Publisher) []models.User {
	users := []models.User{}

	for i := 0; i < numberOfUsers; i++ {
		user := models.User{
			ID:   uint64(i + 1),
			Name: gofakeit.Name(),
			Age:  gofakeit.Number(20, 50),
		}

		userRegisteredPayload, err := json.Marshal(&user)
		if err != nil {
			log.Printf("generateUsers: failed to marshal the payload. error: %v, payload: %v\n", err, user)
			continue
		}

		err = publish(models.UserRegistered, CreateTopic, "generateUsers", userRegisteredPayload, publisher)
		if err == nil {
			users = append(users, user)
		}
	}

	return users
}

func generateItems(numberOfItems int, publisher message.Publisher) []models.Item {
	items := []models.Item{}

	for i := 0; i < numberOfItems; i++ {
		vehicle := gofakeit.Vehicle()
		item := models.Item{
			ID:           uint64(i + 1),
			Title:        fmt.Sprintf("%s %s %d", vehicle.Brand, vehicle.Model, vehicle.Year),
			Manufacturer: vehicle.Brand,
			Origin:       gofakeit.Country(),
		}

		itemCreatedPayload, err := json.Marshal(&item)
		if err != nil {
			log.Printf("generateItems: failed to marshal the payload. error: %v, payload: %v\n", err, item)
			continue
		}

		err = publish(models.ItemCreated, CreateTopic, "generateItems", itemCreatedPayload, publisher)
		if err == nil {
			items = append(items, item)
		}
	}

	return items
}

func publish(eventType, topic, logPrefix string, eventPayload []byte, publisher message.Publisher) error {
	obj := models.Event{
		CreatedAt: time.Now().Unix(),
		Type:      eventType,
		Payload:   eventPayload,
	}
	payload, err := json.Marshal(&obj)
	if err != nil {
		log.Printf("%s: error while marshalling main payload, error %v\n", logPrefix, err)
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(watermill.NewUUID(), msg)
	log.Printf("%s: pushing message ID %s with payload %v\n", logPrefix, msg.UUID, string(msg.Payload))
	return publisher.Publish(topic, msg)
}
