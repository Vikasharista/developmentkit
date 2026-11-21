# Edge & IoT Architecture (Phygital Systems)

Read when the system spans physical and digital touchpoints (in-store devices,
kiosks, sensors, mobile).

- **Edge compute** — Decide what runs on-device vs. cloud. Latency-sensitive and offline-capable logic runs at the edge. Heavy computation and cross-device aggregation runs in the cloud.
- **Connectivity patterns** — Design for intermittent connectivity. Use store-and-forward, local-first data with sync, and conflict resolution strategies (CRDTs, last-write-wins, or domain-specific merge logic).
- **Device fleet management** — OTA (over-the-air) update strategy with staged rollouts, rollback capability, and health reporting. Tools: AWS IoT, Azure IoT Hub, Balena, or custom.
- **Protocol selection** — MQTT for low-bandwidth telemetry, gRPC for device-to-cloud APIs, WebSockets for real-time dashboards, BLE/NFC for proximity interactions.
- **Physical-digital data unification** — Events from physical touchpoints (POS transactions, foot traffic sensors, RFID scans) must flow into the same event backbone as digital events. Design a unified event schema that tags source channel.
- **Offline-first architecture** — For mobile and kiosk apps: local database (SQLite, Realm), background sync queue, conflict resolution, and graceful degradation when the cloud is unreachable.

## Review gate

- [ ] What happens when a device loses connectivity? Is there a local-first fallback?
- [ ] How are devices updated? Is OTA staged with rollback?
- [ ] Are physical and digital events unified into a common event schema?
- [ ] Is device identity and authentication handled (mutual TLS, device certificates)?
- [ ] Is there a device health monitoring and alerting pipeline?
- [ ] Are edge compute resources right-sized for the workload (cost vs. capability)?
