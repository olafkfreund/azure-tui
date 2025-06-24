# Azure DevOps Integration Implementation Plan

## Project Overview

This document outlines the implementation plan for integrating Azure DevOps functionality into the Azure TUI application. The integration will provide comprehensive pipeline management capabilities through a popup interface similar to the existing Terraform module.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Features and Capabilities](#features-and-capabilities)
3. [Implementation Phases](#implementation-phases)
4. [Technical Requirements](#technical-requirements)
5. [API Integration](#api-integration)
6. [User Interface Design](#user-interface-design)
7. [Testing Strategy](#testing-strategy)
8. [Timeline and Milestones](#timeline-and-milestones)

## Architecture Overview

### Design Principles
- **Consistency**: Follow existing patterns used in Terraform, Settings, and Network modules
- **Non-intrusive**: Implement as popup overlay without modifying core TUI structure
- **Modular**: Self-contained module with clear separation of concerns
- **Extensible**: Easy to add new DevOps features in the future

### Component Structure
```
internal/azure/devops/
├── devops.go           # Core DevOps client and API functions
├── pipelines.go        # Pipeline management functions
├── organizations.go    # Organization and project functions
├── builds.go           # Build management functions
├── releases.go         # Release management functions
└── types.go           # DevOps-specific data structures
```

## Features and Capabilities

### Phase 1: Core Pipeline Management
- **Organization & Project Listing**
  - List available Azure DevOps organizations
  - List projects within selected organization
  - Switch between different organizations/projects

- **Pipeline Discovery**
  - List all build pipelines
  - List all release pipelines
  - Filter pipelines by name, status, or recent activity
  - Display pipeline basic information (name, repository, last run)

- **Pipeline Operations**
  - Start/queue new pipeline runs
  - Cancel running pipelines
  - View pipeline run history
  - Display pipeline run status and progress

### Phase 2: Advanced Pipeline Management
- **Pipeline Details**
  - View detailed pipeline configuration
  - Display pipeline variables and parameters
  - Show pipeline triggers and schedules
  - View connected repositories and branches

- **Build Management**
  - View build logs in real-time
  - Download build artifacts
  - View test results and coverage
  - Display build duration and resource usage

- **Release Management**
  - View release pipeline stages
  - Approve/reject release deployments
  - View deployment history
  - Monitor release progress across environments

### Phase 3: Pipeline Creation and Configuration
- **Pipeline Templates**
  - Create pipelines from predefined templates
  - Azure Web App deployment template
  - Container deployment template
  - Infrastructure as Code template
  - Custom YAML pipeline generation

- **Pipeline Configuration**
  - Edit basic pipeline settings
  - Manage pipeline variables
  - Configure build triggers
  - Set up notifications

### Phase 4: Advanced Features
- **Real-time Monitoring**
  - Live pipeline status updates
  - Real-time log streaming
  - Build/release notifications
  - Progress indicators with ETA

- **Analytics and Reporting**
  - Pipeline success/failure rates
  - Average build times
  - Resource utilization metrics
  - Historical trend analysis

## Implementation Phases

### Phase 1: Foundation (Week 1-2)
1. **Create DevOps Module Structure**
   - Set up internal/azure/devops/ package
   - Implement basic Azure DevOps API client
   - Create core data structures and types

2. **Basic Authentication**
   - Implement Personal Access Token (PAT) authentication
   - Add configuration management for DevOps credentials
   - Test basic API connectivity

3. **Organization and Project Management**
   - Implement organization listing
   - Implement project listing within organizations
   - Add organization/project selection interface

### Phase 2: Pipeline Discovery (Week 2-3)
1. **Pipeline Listing**
   - Implement build pipeline discovery
   - Implement release pipeline discovery
   - Add basic filtering and sorting

2. **Pipeline Information Display**
   - Show pipeline basic details
   - Display last run information
   - Add status indicators

3. **Popup Interface Foundation**
   - Create DevOps popup structure
   - Implement navigation between different views
   - Add keyboard shortcuts for DevOps operations

### Phase 3: Pipeline Operations (Week 3-4)
1. **Pipeline Execution**
   - Implement pipeline run triggering
   - Add parameter input for parameterized pipelines
   - Implement run cancellation

2. **Run History and Details**
   - Show pipeline run history
   - Display detailed run information
   - Add run status monitoring

3. **Progress Tracking**
   - Real-time run status updates
   - Progress indicators for running pipelines
   - ETA calculations

### Phase 4: Advanced Features (Week 4-6)
1. **Log Viewing**
   - Implement build log retrieval
   - Add log streaming for active builds
   - Format logs with syntax highlighting

2. **Artifact Management**
   - List build artifacts
   - Implement artifact download
   - Show artifact metadata

3. **Release Management**
   - Release pipeline operations
   - Deployment approvals
   - Environment-specific operations

### Phase 5: Pipeline Creation (Week 6-8)
1. **Template System**
   - Create pipeline template library
   - Implement template-based pipeline creation
   - Add customization options

2. **YAML Generation**
   - Generate Azure Pipelines YAML
   - Template variable substitution
   - Validation and preview

## Technical Requirements

### Dependencies
- **Azure DevOps REST API**: For all DevOps operations
- **Authentication**: Personal Access Token (PAT) support
- **Configuration**: Store DevOps settings in application config
- **Existing TUI Framework**: Bubble Tea, Lipgloss for UI components

### Authentication Methods
1. **Personal Access Token (PAT)**
   - Primary authentication method
   - Stored securely in application configuration
   - Scope: Build (read, execute), Release (read, write, execute, manage)

2. **Environment Variables**
   - `AZURE_DEVOPS_PAT`: Personal Access Token
   - `AZURE_DEVOPS_ORG`: Default organization
   - `AZURE_DEVOPS_PROJECT`: Default project

### API Endpoints
```
Base URL: https://dev.azure.com/{organization}/_apis/

Organizations: https://app.vssps.visualstudio.com/_apis/accounts
Projects: /projects
Pipelines: /pipelines
Builds: /build/builds
Releases: /release/releases
```

## User Interface Design

### Popup Structure
```
┌─ Azure DevOps Manager ─────────────────────────────────┐
│                                                        │
│ Organization: MyOrg          Project: MyProject        │
│                                                        │
│ ┌─ Mode Selection ─────────────────────────────────┐   │
│ │ [1] Organizations  [2] Pipelines  [3] Builds     │   │
│ │ [4] Releases      [5] Templates  [6] Logs       │   │
│ └─────────────────────────────────────────────────┘   │
│                                                        │
│ ┌─ Content Area ───────────────────────────────────┐   │
│ │                                                  │   │
│ │  [Content specific to selected mode]             │   │
│ │                                                  │   │
│ │                                                  │   │
│ │                                                  │   │
│ └─────────────────────────────────────────────────┘   │
│                                                        │
│ DevOps: Navigate: ↑/↓  Select: Enter  Back: Esc       │
└────────────────────────────────────────────────────────┘
```

### Navigation Flow
1. **Entry**: Press 'O' (DevOps) from main interface
2. **Organization Selection**: Choose organization (if multiple)
3. **Project Selection**: Choose project within organization
4. **Mode Selection**: Choose operation mode (pipelines, builds, etc.)
5. **Operations**: Perform specific DevOps operations
6. **Exit**: ESC to go back, 'q' to close popup

### Keyboard Shortcuts
```
Main Interface:
  O          - Open Azure DevOps Manager
  Ctrl+O     - Quick pipeline run (if in DevOps context)

DevOps Popup:
  1-6        - Switch between modes
  Enter      - Select/Execute action
  Space      - Toggle selection (where applicable)
  r          - Refresh current view
  n          - New/Create operation
  s          - Start/Queue pipeline
  x          - Cancel/Stop operation
  l          - View logs
  h          - View history
  ↑/↓        - Navigate items
  ←/→        - Navigate between columns/tabs
  Esc        - Go back/Close popup
  ?          - Show help for current mode
```

## API Integration

### Authentication Configuration
```go
type DevOpsConfig struct {
    PersonalAccessToken string `json:"personal_access_token"`
    Organization        string `json:"organization"`
    Project            string `json:"project"`
    BaseURL            string `json:"base_url"`
}
```

### Core API Client
```go
type DevOpsClient struct {
    baseURL      string
    organization string
    project      string
    token        string
    httpClient   *http.Client
}

func NewDevOpsClient(config DevOpsConfig) *DevOpsClient
func (c *DevOpsClient) ListOrganizations() ([]Organization, error)
func (c *DevOpsClient) ListProjects() ([]Project, error)
func (c *DevOpsClient) ListPipelines() ([]Pipeline, error)
func (c *DevOpsClient) RunPipeline(pipelineId int, parameters map[string]string) (*PipelineRun, error)
func (c *DevOpsClient) CancelPipelineRun(runId int) error
func (c *DevOpsClient) GetPipelineRun(runId int) (*PipelineRun, error)
func (c *DevOpsClient) GetBuildLogs(buildId int) ([]LogEntry, error)
```

### Data Structures
```go
type Organization struct {
    ID          string `json:"accountId"`
    Name        string `json:"accountName"`
    URL         string `json:"accountUri"`
    Description string `json:"description"`
}

type Project struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    URL         string `json:"url"`
    State       string `json:"state"`
}

type Pipeline struct {
    ID           int    `json:"id"`
    Name         string `json:"name"`
    Path         string `json:"path"`
    Repository   Repository `json:"repository"`
    LastRun      *PipelineRun `json:"lastRun,omitempty"`
    Status       string `json:"status"`
}

type PipelineRun struct {
    ID          int       `json:"id"`
    BuildNumber string    `json:"buildNumber"`
    Status      string    `json:"status"`
    Result      string    `json:"result"`
    StartTime   time.Time `json:"startTime"`
    FinishTime  time.Time `json:"finishTime"`
    Duration    time.Duration
    RequestedBy User      `json:"requestedBy"`
}
```

## Testing Strategy

### Unit Tests
- **API Client Tests**: Mock Azure DevOps API responses
- **Data Structure Tests**: Validate JSON parsing and data mapping
- **Business Logic Tests**: Test pipeline operations and state management

### Integration Tests
- **API Integration**: Test against Azure DevOps test organization
- **Authentication Tests**: Validate PAT authentication flow
- **End-to-End Scenarios**: Complete pipeline management workflows

### Manual Testing
- **UI/UX Testing**: Verify popup interface and navigation
- **Performance Testing**: Test with large numbers of pipelines
- **Error Handling**: Test network failures and API errors

## Timeline and Milestones

### Week 1-2: Foundation
- [ ] Create DevOps module structure
- [ ] Implement basic Azure DevOps API client
- [ ] Add authentication and configuration management
- [ ] Create core data structures
- [ ] Implement organization and project listing

### Week 2-3: Pipeline Discovery
- [ ] Implement pipeline listing and filtering
- [ ] Create popup interface foundation
- [ ] Add basic navigation and keyboard shortcuts
- [ ] Display pipeline information and status

### Week 3-4: Pipeline Operations
- [ ] Implement pipeline run triggering
- [ ] Add run history and details viewing
- [ ] Create real-time status monitoring
- [ ] Add run cancellation functionality

### Week 4-6: Advanced Features
- [ ] Implement log viewing and streaming
- [ ] Add artifact management
- [ ] Create release pipeline operations
- [ ] Implement deployment approvals

### Week 6-8: Pipeline Creation
- [ ] Create pipeline template system
- [ ] Implement YAML generation
- [ ] Add template customization
- [ ] Create pipeline validation

## Success Criteria

### Functional Requirements
- ✅ Successfully authenticate with Azure DevOps using PAT
- ✅ List and navigate organizations and projects
- ✅ View and manage build and release pipelines
- ✅ Start, stop, and monitor pipeline runs
- ✅ View build logs and download artifacts
- ✅ Create pipelines from templates

### Performance Requirements
- ⚡ Pipeline list loads within 2 seconds
- ⚡ Real-time status updates every 5 seconds
- ⚡ Log streaming with minimal latency
- ⚡ Responsive UI with smooth navigation

### User Experience Requirements
- 🎨 Consistent interface matching existing TUI patterns
- 🎨 Intuitive keyboard navigation
- 🎨 Clear status indicators and progress feedback
- 🎨 Helpful error messages and recovery options

## Future Enhancements

### Advanced Pipeline Features
- Pipeline as Code (YAML) editing within TUI
- Advanced pipeline analytics and reporting
- Integration with Azure Resource Manager
- Multi-stage pipeline visualization

### DevOps Ecosystem Integration
- Azure Boards integration (work items, backlogs)
- Azure Repos integration (branch management, pull requests)
- Azure Test Plans integration
- Azure Artifacts integration

### Automation and Scripting
- Batch pipeline operations
- Automated deployment workflows
- Pipeline health monitoring
- Custom notification systems

## Conclusion

This implementation plan provides a comprehensive roadmap for integrating Azure DevOps functionality into the Azure TUI. By following the established patterns and maintaining consistency with existing features, we can deliver a powerful and user-friendly DevOps management interface that enhances the overall Azure TUI experience.

The phased approach ensures steady progress while allowing for feedback and iteration at each stage. The focus on core pipeline management in early phases provides immediate value, while advanced features and pipeline creation capabilities extend the tool's usefulness for DevOps workflows.
