// Package svc provides a framework for a typical microservice, providing health checking and monitoring
// whilst possible to write a service without svc, it is not recommended as you are likely to diverge from
// organisation standards.
package service

// TODO: this test doesn't work well, so needs a rethink. It tries to open port 8080 which is sometimes not available.
// The code could be edited so the port was configurable, using 0 would mean a random available port would be chosen

// func TestService_Healthcheck(t *testing.T) {
// 	if testing.Short() {
// 		t.Skipf("skipping e2e test")
// 	}

// 	testCtx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	svc, ctx := New(testCtx, "test-name", "test-env", "", 0, prometheus.NewRegistry(), zap.NewNop())
// 	assert.Nil(t, ctx.Err())
// 	assert.Equal(t, "test-name", svc.Name)

// 	resp, err := svc.health.Check(testCtx, &grpc_health_v1.HealthCheckRequest{})
// 	if err != nil {
// 		t.Errorf("error when checking health: %v", err)
// 	}
// 	assert.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING.String(), resp.Status.String())

// 	go func() {
// 		svc.Run()
// 	}()
// 	<-time.After(100 * time.Millisecond)
// 	resp, err = svc.health.Check(testCtx, &grpc_health_v1.HealthCheckRequest{})
// 	if err != nil {
// 		t.Errorf("error when checking health: %v", err)
// 	}
// 	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING.String(), resp.Status.String())

// 	// fake a shutdown signal
// 	svc.cancel()
// 	<-time.After(100 * time.Millisecond)
// 	resp, err = svc.health.Check(testCtx, &grpc_health_v1.HealthCheckRequest{})
// 	assert.Nil(t, err)
// 	assert.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING.String(), resp.Status.String())
// }
