package service

import (
	"context"
	"fmt"
	"strings"

	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"
	voicerecognitionv1 "voice-service/pkg/pb/voice_recognition/v1"

	"google.golang.org/grpc"
)

type ScenarioClient interface {
	ListVoiceCommands(ctx context.Context, in *scenariov1.ListVoiceCommandsRequest, opts ...grpc.CallOption) (*scenariov1.ListVoiceCommandsResponse, error)
	ExecuteVoiceCommand(ctx context.Context, in *scenariov1.ExecuteVoiceCommandRequest, opts ...grpc.CallOption) (*scenariov1.ExecuteVoiceCommandResponse, error)
}

type RecognitionClient interface {
	GetAudio(ctx context.Context, in *voicerecognitionv1.GetAudioRequest, opts ...grpc.CallOption) (*voicerecognitionv1.GetAudioResponse, error)
}

type Service struct {
	scenario    ScenarioClient
	recognition RecognitionClient
}

func New(scenario ScenarioClient, recognition RecognitionClient) *Service {
	return &Service{
		scenario:    scenario,
		recognition: recognition,
	}
}

func (s *Service) ParseAndExecute(ctx context.Context, req *voicev1.ParseVoiceCommandRequest) (*voicev1.ParsedVoiceCommand, error) {
	if req.GetEdgeId() == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	if len(req.GetAudio()) == 0 {
		return nil, fmt.Errorf("audio is required")
	}

	source := req.GetSource()
	if source == "" {
		source = "voice-service"
	}

	commandsResponse, err := s.scenario.ListVoiceCommands(ctx, &scenariov1.ListVoiceCommandsRequest{
		EdgeId: req.GetEdgeId(),
		RoomId: req.GetRoomId(),
	})
	if err != nil {
		return nil, fmt.Errorf("list voice commands: %w", err)
	}
	if len(commandsResponse.GetCommands()) == 0 {
		return &voicev1.ParsedVoiceCommand{
			Intent:          "execute_voice_scenario",
			RoomId:          req.GetRoomId(),
			ExecutionStatus: "no_available_commands",
		}, nil
	}

	names := make([]string, 0, len(commandsResponse.GetCommands()))
	index := make(map[string]*scenariov1.VoiceCommand, len(commandsResponse.GetCommands()))
	for _, command := range commandsResponse.GetCommands() {
		if command.GetCommandName() == "" {
			continue
		}
		names = append(names, command.GetCommandName())
		index[normalize(command.GetCommandName())] = command
	}

	recognized, err := s.recognize(ctx, req.GetAudio(), names)
	if err != nil {
		return nil, err
	}
	if recognized == "" {
		return &voicev1.ParsedVoiceCommand{
			Intent:            "execute_voice_scenario",
			RoomId:            req.GetRoomId(),
			ExecutionStatus:   "no_match",
			RecognizedCommand: "",
		}, nil
	}

	execResponse, err := s.scenario.ExecuteVoiceCommand(ctx, &scenariov1.ExecuteVoiceCommandRequest{
		EdgeId:         req.GetEdgeId(),
		RoomId:         req.GetRoomId(),
		CommandName:    recognized,
		Source:         source,
		DeferExecution: req.GetDeferExecution(),
	})
	if err != nil {
		return nil, fmt.Errorf("execute voice command: %w", err)
	}

	selected := index[normalize(recognized)]
	result := &voicev1.ParsedVoiceCommand{
		CommandId:         execResponse.GetDecision().GetDecisionId(),
		Intent:            "execute_voice_scenario",
		RoomId:            req.GetRoomId(),
		RecognizedCommand: recognized,
		ScenarioId:        execResponse.GetScenarioId(),
		ScenarioName:      execResponse.GetScenarioName(),
		ExecutionStatus:   execResponse.GetStatus(),
		DecisionId:        execResponse.GetDecision().GetDecisionId(),
	}
	if selected != nil {
		result.DeviceId = selected.GetDeviceId()
		result.EntityId = selected.GetEntityId()
		result.TargetState = selected.GetTargetState()
		result.Confidence = 0.99
	}
	return result, nil
}

func (s *Service) ExecutePhrase(ctx context.Context, phrase, edgeID, roomID, source string) (*voicev1.ParsedVoiceCommand, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	if strings.TrimSpace(phrase) == "" {
		return nil, fmt.Errorf("phrase is required")
	}
	if source == "" {
		source = "voice-service"
	}

	execResponse, err := s.scenario.ExecuteVoiceCommand(ctx, &scenariov1.ExecuteVoiceCommandRequest{
		EdgeId:         edgeID,
		RoomId:         roomID,
		CommandName:    phrase,
		Source:         source,
		DeferExecution: false,
	})
	if err != nil {
		return nil, fmt.Errorf("execute phrase: %w", err)
	}

	return &voicev1.ParsedVoiceCommand{
		CommandId:         execResponse.GetDecision().GetDecisionId(),
		Intent:            "execute_voice_scenario",
		RoomId:            roomID,
		RecognizedCommand: phrase,
		ScenarioId:        execResponse.GetScenarioId(),
		ScenarioName:      execResponse.GetScenarioName(),
		ExecutionStatus:   execResponse.GetStatus(),
		DecisionId:        execResponse.GetDecision().GetDecisionId(),
	}, nil
}

func (s *Service) recognize(ctx context.Context, audio []byte, commands []string) (string, error) {
	if len(commands) == 0 {
		return "", nil
	}

	if s.recognition == nil {
		return "", fmt.Errorf("voice recognition client is not configured")
	}

	request := &voicerecognitionv1.GetAudioRequest{
		Chunk: audio,
	}
	for _, command := range commands {
		request.Commands = append(request.Commands, &voicerecognitionv1.GetAudioRequest_Command{Name: command})
	}
	response, err := s.recognition.GetAudio(ctx, request)
	if err != nil {
		return "", fmt.Errorf("recognize audio: %w", err)
	}
	return response.GetCommand(), nil
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
