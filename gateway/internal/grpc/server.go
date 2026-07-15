package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/hotel-voice-agent/gateway/internal/cache"
	"github.com/hotel-voice-agent/gateway/internal/db"
	pb "github.com/hotel-voice-agent/gateway/proto"
	"google.golang.org/grpc"
)

type HotelStateServer struct {
	pb.UnimplementedHotelStateServiceServer
	repo db.BookingRepository
}

func NewHotelStateServer(repo db.BookingRepository) *HotelStateServer {
	return &HotelStateServer{repo: repo}
}

func (s *HotelStateServer) CheckAvailability(ctx context.Context, req *pb.AvailabilityRequest) (*pb.AvailabilityResponse, error) {
	available, err := cache.GetAvailableRooms(s.repo, req.RoomType)
	if err != nil {
		return &pb.AvailabilityResponse{
			IsAvailable:  false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.AvailabilityResponse{
		IsAvailable:    available > 0,
		AvailableCount: int32(available),
	}, nil
}

func (s *HotelStateServer) InitiateCheckout(ctx context.Context, req *pb.CheckoutRequest) (*pb.CheckoutResponse, error) {
	// 1. Create pending booking in DB
	bookingID, err := s.repo.CreateBooking(req.GuestName, req.RoomType, int(req.Nights))
	if err != nil {
		return &pb.CheckoutResponse{ErrorMessage: err.Error()}, nil
	}

	// 2. Invalidate cache since a room is now held
	cache.InvalidateAvailability(req.RoomType)

	// 3. TODO: Call Xendit API to generate invoice
	// For now, we mock the Xendit response
	mockInvoiceID := fmt.Sprintf("inv_%d", bookingID)
	mockInvoiceURL := fmt.Sprintf("https://checkout.xendit.co/web/%s", mockInvoiceID)

	// 4. Update DB with invoice ID (this would be done via repo method)
	// err = s.repo.AssociateInvoice(bookingID, mockInvoiceID)
	
	log.Printf("Initiated checkout for booking %d, Invoice: %s", bookingID, mockInvoiceID)

	return &pb.CheckoutResponse{
		InvoiceUrl: mockInvoiceURL,
		InvoiceId:  mockInvoiceID,
	}, nil
}

func StartServer(port int, repo db.BookingRepository) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	pb.RegisterHotelStateServiceServer(s, NewHotelStateServer(repo))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
