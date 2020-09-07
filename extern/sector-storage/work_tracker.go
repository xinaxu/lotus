package sectorstorage

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-storage/storage"

	"github.com/filecoin-project/lotus/extern/sector-storage/sealtasks"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
)

type workTracker struct {
	lk sync.Mutex

	done    map[storiface.CallID]struct{}
	running map[storiface.CallID]storiface.WorkerJob

	// TODO: done, aggregate stats, queue stats, scheduler feedback
}

// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
// TODO: CALL THIS!
func (wt *workTracker) onDone(callID storiface.CallID) {
	wt.lk.Lock()
	defer wt.lk.Unlock()

	_, ok := wt.running[callID]
	if !ok {
		wt.done[callID] = struct{}{}
		return
	}

	delete(wt.running, callID)
}

func (wt *workTracker) track(sid abi.SectorID, task sealtasks.TaskType) func(storiface.CallID, error) (storiface.CallID, error) {
	return func(callID storiface.CallID, err error) (storiface.CallID, error) {
		if err != nil {
			return callID, err
		}

		wt.lk.Lock()
		defer wt.lk.Unlock()

		_, done := wt.done[callID]
		if done {
			delete(wt.done, callID)
			return callID, err
		}

		wt.running[callID] = storiface.WorkerJob{
			ID:     callID,
			Sector: sid,
			Task:   task,
			Start:  time.Now(),
		}

		return callID, err
	}
}

func (wt *workTracker) worker(w Worker) Worker {
	return &trackedWorker{
		Worker:  w,
		tracker: wt,
	}
}

func (wt *workTracker) Running() []storiface.WorkerJob {
	wt.lk.Lock()
	defer wt.lk.Unlock()

	out := make([]storiface.WorkerJob, 0, len(wt.running))
	for _, job := range wt.running {
		out = append(out, job)
	}

	return out
}

type trackedWorker struct {
	Worker

	tracker *workTracker
}

func (t *trackedWorker) SealPreCommit1(ctx context.Context, sector abi.SectorID, ticket abi.SealRandomness, pieces []abi.PieceInfo) (storiface.CallID, error) {
	return t.tracker.track(sector, sealtasks.TTPreCommit1)(t.Worker.SealPreCommit1(ctx, sector, ticket, pieces))
}

func (t *trackedWorker) SealPreCommit2(ctx context.Context, sector abi.SectorID, pc1o storage.PreCommit1Out) (storiface.CallID, error) {
	return t.tracker.track(sector, sealtasks.TTPreCommit2)(t.Worker.SealPreCommit2(ctx, sector, pc1o))
}

func (t *trackedWorker) SealCommit1(ctx context.Context, sector abi.SectorID, ticket abi.SealRandomness, seed abi.InteractiveSealRandomness, pieces []abi.PieceInfo, cids storage.SectorCids) (storiface.CallID, error) {
	return t.tracker.track(sector, sealtasks.TTCommit1)(t.Worker.SealCommit1(ctx, sector, ticket, seed, pieces, cids))
}

func (t *trackedWorker) SealCommit2(ctx context.Context, sector abi.SectorID, c1o storage.Commit1Out) (storiface.CallID, error) {
	return t.tracker.track(sector, sealtasks.TTCommit2)(t.Worker.SealCommit2(ctx, sector, c1o))
}

func (t *trackedWorker) FinalizeSector(ctx context.Context, sector abi.SectorID, keepUnsealed []storage.Range) (storiface.CallID, error) {
	return t.tracker.track(sector, sealtasks.TTFinalize)(t.Worker.FinalizeSector(ctx, sector, keepUnsealed))
}

func (t *trackedWorker) AddPiece(ctx context.Context, sector abi.SectorID, pieceSizes []abi.UnpaddedPieceSize, newPieceSize abi.UnpaddedPieceSize, pieceData storage.Data) (storiface.CallID, error) {
	return t.tracker.track(sector, sealtasks.TTAddPiece)(t.Worker.AddPiece(ctx, sector, pieceSizes, newPieceSize, pieceData))
}

func (t *trackedWorker) Fetch(ctx context.Context, s abi.SectorID, ft storiface.SectorFileType, ptype storiface.PathType, am storiface.AcquireMode) (storiface.CallID, error) {
	return t.tracker.track(s, sealtasks.TTFetch)(t.Worker.Fetch(ctx, s, ft, ptype, am))
}

func (t *trackedWorker) UnsealPiece(ctx context.Context, id abi.SectorID, index storiface.UnpaddedByteIndex, size abi.UnpaddedPieceSize, randomness abi.SealRandomness, cid cid.Cid) (storiface.CallID, error) {
	return t.tracker.track(id, sealtasks.TTUnseal)(t.Worker.UnsealPiece(ctx, id, index, size, randomness, cid))
}

func (t *trackedWorker) ReadPiece(ctx context.Context, writer io.Writer, id abi.SectorID, index storiface.UnpaddedByteIndex, size abi.UnpaddedPieceSize) (storiface.CallID, error) {
	return t.tracker.track(id, sealtasks.TTReadUnsealed)(t.Worker.ReadPiece(ctx, writer, id, index, size))
}

var _ Worker = &trackedWorker{}
