package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	shared "github.com/qtopie/homa/internal/assistant/plugins/copilot/shared"
)

type EtcdStore struct {
    cli       *clientv3.Client
    maxItems  int
    ttlSeconds int64
}

func NewEtcdStore(endpoints []string, maxItems int, ttlSeconds int64) (*EtcdStore, error) {
    if len(endpoints) == 0 {
        endpoints = []string{"localhost:2379"}
    }
    cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
    if err != nil {
        return nil, err
    }
    return &EtcdStore{cli: cli, maxItems: maxItems, ttlSeconds: ttlSeconds}, nil
}

func (s *EtcdStore) key(sessionID string) string {
    return fmt.Sprintf("/sessions/%s/history", sessionID)
}

// AppendHistory appends a message to the session history and trims to maxItems.
func (s *EtcdStore) AppendHistory(ctx context.Context, sessionID string, msg shared.Message) error {
    key := s.key(sessionID)
    for {
        getResp, err := s.cli.Get(ctx, key)
        if err != nil {
            return err
        }

        var hist []shared.Message
        var modRev int64 = 0
        if len(getResp.Kvs) > 0 {
            modRev = getResp.Kvs[0].ModRevision
            if err := json.Unmarshal(getResp.Kvs[0].Value, &hist); err != nil {
                hist = nil
            }
        }

        hist = append(hist, msg)
        if len(hist) > s.maxItems {
            hist = hist[len(hist)-s.maxItems:]
        }
        data, err := json.Marshal(hist)
        if err != nil {
            return err
        }

        var putOpts []clientv3.OpOption
        if s.ttlSeconds > 0 {
            leaseResp, err := s.cli.Grant(ctx, s.ttlSeconds)
            if err != nil {
                return err
            }
            putOpts = append(putOpts, clientv3.WithLease(leaseResp.ID))
        }

        var cmp clientv3.Cmp
        if modRev == 0 {
            cmp = clientv3.Compare(clientv3.Version(key), "=", 0)
        } else {
            cmp = clientv3.Compare(clientv3.ModRevision(key), "=", modRev)
        }

        txn := s.cli.Txn(ctx).If(cmp).Then(clientv3.OpPut(key, string(data), putOpts...))
        txnResp, err := txn.Commit()
        if err != nil {
            return err
        }
        if txnResp.Succeeded {
            return nil
        }
        // compare failed, retry
        time.Sleep(10 * time.Millisecond)
    }
}

// GetHistory returns up to maxItems recent messages for a session.
func (s *EtcdStore) GetHistory(ctx context.Context, sessionID string) ([]shared.Message, error) {
    key := s.key(sessionID)
    getResp, err := s.cli.Get(ctx, key)
    if err != nil {
        return nil, err
    }
    if len(getResp.Kvs) == 0 {
        return nil, nil
    }
    var hist []shared.Message
    if err := json.Unmarshal(getResp.Kvs[0].Value, &hist); err != nil {
        return nil, err
    }
    if len(hist) > s.maxItems {
        hist = hist[len(hist)-s.maxItems:]
    }
    return hist, nil
}

// Close closes underlying etcd client.
func (s *EtcdStore) Close() error {
    if s.cli == nil {
        return nil
    }
    return s.cli.Close()
}
