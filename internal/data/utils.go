package data

import (
	"strings"

	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// orEmpty возвращает строку или пустую, если nil
func orEmpty(s *string) string {
	if s == nil || *s == "" {
		return ""
	}
	return *s
}

// escapeTabs заменяет управляющие символы на пробелы
func escapeTabs(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// shareTypeToString — ShareType → строка
func shareTypeToString(t pb.ShareType) string {
	switch t {
	case pb.ShareType_SHARE_TYPE_COMMON:
		return "common"
	case pb.ShareType_SHARE_TYPE_PREFERRED:
		return "preferred"
	default:
		return ""
	}
}

// tradingStatusToString преобразует enum в читаемую строку
func tradingStatusToString(status pb.SecurityTradingStatus) string {
	switch status {
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_UNSPECIFIED:
		return "unspecified"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_NOT_AVAILABLE_FOR_TRADING:
		return "not_available"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_OPENING_PERIOD:
		return "opening_period"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_CLOSING_PERIOD:
		return "closing_period"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_BREAK_IN_TRADING:
		return "break_in_trading"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_NORMAL_TRADING:
		return "normal_trading"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_CLOSING_AUCTION:
		return "closing_auction"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_DARK_POOL_AUCTION:
		return "dark_pool_auction"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_DISCRETE_AUCTION:
		return "discrete_auction"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_OPENING_AUCTION_PERIOD:
		return "opening_auction_period"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_TRADING_AT_CLOSING_AUCTION_PRICE:
		return "trading_at_close_price"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_SESSION_ASSIGNED:
		return "session_assigned"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_SESSION_CLOSE:
		return "session_close"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_SESSION_OPEN:
		return "session_open"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_DEALER_NORMAL_TRADING:
		return "dealer_normal_trading"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_DEALER_BREAK_IN_TRADING:
		return "dealer_break"
	case pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_DEALER_NOT_AVAILABLE_FOR_TRADING:
		return "dealer_not_available"
	default:
		return "unknown"
	}
}
