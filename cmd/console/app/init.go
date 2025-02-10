package app

import (
	"context"
	"go.uber.org/zap"
	"time"
)

func (s *ConsoleServer) initSystem() (err error) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	// defer cancel()

	zap.L().Info("initSystem....")
	defer func(t1 time.Time) {
		if err != nil {
			zap.L().Info("initSystem exception！！！", zap.Duration("since", time.Since(t1)), zap.Error(err))
		} else {
			zap.L().Info("initSystem done!", zap.Duration("since", time.Since(t1)), zap.Error(err))
		}
	}(time.Now())

	s.DeleteTaskMonitor.Start(context.Background(), time.Second*10)

	return err
}
