// Copyright 2018-2020 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dmap

import (
	"github.com/buraksezer/olric/internal/cluster/partitions"
	"github.com/buraksezer/olric/internal/protocol"
)

func (s *Service) exGetOperation(w, r protocol.EncodeDecoder) {
	req := r.(*protocol.DMapMessage)
	dm, err := s.LoadDMap(req.DMap())
	if err != nil {
		errorResponse(w, err)
	}

	entry, err := dm.get(req.DMap(), req.Key())
	if err != nil {
		errorResponse(w, err)
		return
	}
	w.SetStatus(protocol.StatusOK)
	w.SetValue(entry.Encode())
}

func (s *Service) getBackupOperation(w, r protocol.EncodeDecoder) {
	req := r.(*protocol.DMapMessage)
	dm, err := s.LoadDMap(req.DMap())
	if err != nil {
		errorResponse(w, err)
	}

	hkey := partitions.HKey(req.DMap(), req.Key())
	f, err := dm.getFragment(req.DMap(), hkey, partitions.BACKUP)
	if err != nil {
		errorResponse(w, err)
		return
	}
	f.RLock()
	defer f.RUnlock()
	entry, err := f.storage.Get(hkey)
	if err != nil {
		errorResponse(w, err)
		return
	}
	if isKeyExpired(entry.TTL()) {
		errorResponse(w, ErrKeyNotFound)
		return
	}
	w.SetStatus(protocol.StatusOK)
	w.SetValue(entry.Encode())
}

func (s *Service) getPrevOperation(w, r protocol.EncodeDecoder) {
	req := r.(*protocol.DMapMessage)
	dm, err := s.LoadDMap(req.DMap())
	if err != nil {
		errorResponse(w, err)
	}

	hkey := partitions.HKey(req.DMap(), req.Key())
	f, err := dm.getFragment(req.DMap(), hkey, partitions.PRIMARY)
	if err != nil {
		errorResponse(w, err)
	}

	f.RLock()
	defer f.RUnlock()

	entry, err := f.storage.Get(hkey)
	if err != nil {
		errorResponse(w, err)
		return
	}

	if isKeyExpired(entry.TTL()) {
		errorResponse(w, ErrKeyNotFound)
		return
	}

	w.SetStatus(protocol.StatusOK)
	w.SetValue(entry.Encode())
}
