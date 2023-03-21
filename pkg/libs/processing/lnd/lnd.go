package lnd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/utils"
	"os"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type (
	macaroonCredential struct {
		macaroon []byte
	}
	InvoiceRequest struct {
		Amount int64
		Expiry int64
		Memo   string
	}
	InvoiceResponse struct {
		Invoice *lnrpc.AddInvoiceResponse
	}
	CheckInvoiceRequest struct {
		RHASH []byte
	}
)

func getClient() *grpc.ClientConn {
	cert, err := os.ReadFile(config.Config.Processing.Lnd.Certificate)
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		panic("failed to append certificate")
	}
	transportCredentials := credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	})

	macaroonBytes, err := os.ReadFile(config.Config.Processing.Lnd.Macaroon)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(
			NewMacaroonCredential(macaroonBytes),
		),
	}

	conn, err := grpc.Dial(config.Config.Processing.Lnd.Host, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

func CreateInvoice(invoiceReq *InvoiceRequest) (InvoiceResponse, error) {
	conn := getClient()
	defer conn.Close()
	client := lnrpc.NewLightningClient(conn)

	invoice := &lnrpc.Invoice{
		Memo:      invoiceReq.Memo,
		ValueMsat: invoiceReq.Amount,
		Expiry:    invoiceReq.Expiry,
	}

	invoiceResponse, err := client.AddInvoice(context.Background(), invoice)
	if err != nil {
		return InvoiceResponse{}, err
	}

	response := InvoiceResponse{Invoice: invoiceResponse}
	return response, nil
}

func CheckInvoicePaid(checkInvoice *CheckInvoiceRequest) bool {
	conn := getClient()
	defer conn.Close()
	client := lnrpc.NewLightningClient(conn)
	query := &lnrpc.PaymentHash{
		RHash: checkInvoice.RHASH,
	}
	lookupInvoice, err := client.LookupInvoice(context.Background(), query)
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("LND: Can't check invoice status:\n%v", err),
			Level:    "ERROR",
		})
		return false
	}
	return lookupInvoice.State == 1
}

func listInvoices() {
	conn := getClient()
	defer conn.Close()
	client := lnrpc.NewLightningClient(conn)

	test := &lnrpc.ListInvoiceRequest{
		PendingOnly:    true,
		IndexOffset:    3,
		NumMaxInvoices: 0,
	}

	r, err := client.ListInvoices(context.Background(), test)
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
}

// NewMacaroonCredential creates a new macaroon credential from the given bytes
func NewMacaroonCredential(macaroonBytes []byte) credentials.PerRPCCredentials {
	return macaroonCredential{
		macaroon: macaroonBytes,
	}
}

func (c macaroonCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"macaroon": fmt.Sprintf("%x", c.macaroon),
	}, nil
}

func (c macaroonCredential) RequireTransportSecurity() bool {
	return true
}
