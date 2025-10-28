package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/model"
	"EventBooker/internal/repository"
)

type SchedulerService struct {
	bookingRepo repository.BookingRepository
}

func NewSchedulerService(bookingRepo repository.BookingRepository) *SchedulerService {
	return &SchedulerService{bookingRepo: bookingRepo}
}

func (s *SchedulerService) Start(ctx context.Context, interval time.Duration, msgCh chan<- RetryMessage) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expiredBooking, err := s.bookingRepo.GetExpiredBooking(ctx)
			if err != nil {
				zlog.Logger.Error().Msgf("serviceSchedulerService.Start error: %v", err)
			}

			for _, b := range expiredBooking {
				select {
				case <-ctx.Done():
					return
				case msgCh <- buildMessage(b):
				}
			}

			zlog.Logger.Info().Msgf("Start delete expired booking")
			s.deleteExpiredBooking(ctx)
		}
	}
}

func (s *SchedulerService) deleteExpiredBooking(ctx context.Context) {
	err := s.bookingRepo.DeleteExpiredBooking(ctx)
	if err != nil {
		zlog.Logger.Error().Msgf("serviceSchedulerService.deleteExpiredBooking error: %v", err)
	}
}

func buildMessage(b model.BookingGetForTG) RetryMessage {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Бронирование %d отменено.\n", b.ID))
	builder.WriteString(fmt.Sprintf("Событие: %s\n", b.TitleEvent))
	builder.WriteString(fmt.Sprintf("Время события: %s\n", b.EventDate))
	return RetryMessage{
		ChatID: b.TgChatID,
		Text:   builder.String(),
	}
}
