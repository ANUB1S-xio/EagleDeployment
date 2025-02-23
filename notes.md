
Web Interface with API Endpoints
Login & Dashboard Pages:
  /: Default login page.
  /dashboard: Main interface after successful authentication.

Playbook Management Web Pages:
  /execute: GUI for selecting and executing playbooks.
  /list: GUI for viewing and editing YAML playbooks.
  API Endpoints:

  GET /api/list_playbooks: Retrieves available YAML files from the /playbooks directory.
  POST /api/execute: Executes selected YAML playbooks, integrated with backend execution logic.

Backend Execution Integration (executor.go)
Wrapper Function Added:

ExecuteYAML(playbookPath, targetHosts)
Loads the YAML playbook.
Injects current inventory data.
Executes tasks concurrently across hosts.
Original executor methods preserved:

ExecuteRemote(): Remote SSH execution.
ExecuteLocal(): Local command execution.
ExecuteConcurrently(): Concurrent task execution.

Inventory Management Integration (inventory.go)
Automatically injects real-time inventory data (hosts, credentials) into YAML playbooks.
Ensures playbooks are dynamically updated with accurate system information before execution.

++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
Fixes and Enhancements:

Inventory Injection Fix:
  Problem: Playbooks executed without current inventory, leading to outdated host or credential errors.
  Solution: Integrated inventory injection logic (inventory.InjectInventoryIntoPlaybook) directly into the execution workflow.

API Endpoint Implementation:
  Problem: No structured API existed for frontend/backend communication.
  Solution: Clearly defined RESTful API endpoints (/api/list_playbooks, /api/execute) for robust integration.

++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

Final Workflow:
  Admin accesses localhost:8742.
  Login Authentication: Admin credentials verified.
  Dashboard Loaded: Presents clear options:
  List YAML: Admin selects, views, and edits YAML playbooks.
  Execute YAML: Admin chooses a YAML file, triggers backend execution via API call.
  Backend Execution: Tasks defined in YAML run concurrently across specified hosts, with full inventory integration.


EagleDeployment/
├── main.go                  # CLI entry point & menu logic
├── web/
│   ├── web.go               # Webserver (dynamic ports & API handlers)
│   ├── templates/
│   │   ├── login.html
│   │   ├── dashboard.html
│   │   ├── execute.html
│   │   └── list.html
│   └── static/
│       └── css/
│           ├── styles.css
│           ├── login.css
│           └── dashboard.css
├── executor/
│   └── executor.go          # Backend execution logic
├── inventory/
│   ├── inventory.go         # Inventory management logic
│   └── inventory.yaml       # Hosts & credentials
├── playbooks/
│   └── (yaml playbooks)
└── sshutils/
|   └── (ssh connection utilities)
├── osdetect/
│   └── os-detect.go
