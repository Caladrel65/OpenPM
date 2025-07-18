# OpenPM
OpenPM is a collection of technologies that creates a lightweight Project Manager to assist with solo developers or small groups.

The goal is to create a stack where the "PM" will reach out to the developer on a schedule, asking about the status of tasks to help improve accountability.

My current plan (as of the initial ideation of this project) is to use a locally-run LLM with a system prompt to work as a PM. This LLM will use RAG or an MCP server to retrieve data on a GitHub repository, where it will determine current tasks and decide what may need to be checked in on. It should also utilize either a daily notes system (so the developer can note their progress) or maintain conversation history for the latest reported status of the task. The LLM will be prompted on some cadence, possibly daily towards EOD during the week, to ask for updates and try to provide direction if any is needed. The response will be sent to the developer, possibly via webhook (so various systems such as e.g. a Discord bot could report it), which the developer can then respond to. Once all responses are finished, it will not respond until the next appropriate time (likely the next scheduled time).

Future improvements could include a custom scheduler, which allows for e.g. weekend work if prompted ("I'll have time this Saturday, can you check in on me then?").

## Scheduling and Automation

This application is designed to be run on a schedule to provide regular check-ins. Here’s how to set it up on Windows using the Task Scheduler.

### 1. Compile the Application

First, you need to build the Go application into a standalone executable. Run the following command in your project directory:

```bash
go build -o OpenPM.exe
```

This will create `OpenPM.exe` in your project folder.

### 2. Set up Discord Integration (Optional)

To receive notifications in Discord, you need to create a webhook.

1.  **Create a Webhook:**
    *   In your Discord server, go to `Server Settings` > `Integrations` > `Webhooks`.
    *   Click `New Webhook`.
    *   Give it a name (e.g., "Project Manager") and choose the channel it should post to.
    *   Copy the `Webhook URL`.

2.  **Set Environment Variable:**
    *   You need to set the `DISCORD_WEBHOOK_URL` environment variable to the URL you just copied. You can do this in your system settings or in the terminal session where you run the application.
    *   For example, in PowerShell:
        ```powershell
        $env:DISCORD_WEBHOOK_URL="your_webhook_url_here"
        ```

### 3. Schedule the Task

You can use the `schtasks.exe` command-line tool to create a new scheduled task. Here is an example that schedules the `OpenPM.exe` to run every weekday at 4:00 PM (16:00):

```bash
schtasks /create /tn "OpenPM Daily Check-in" /tr "C:\path	o\your\project\OpenPM.exe" /sc DAILY /mo 1 /d MON,TUE,WED,THU,FRI /st 16:00
```

**Important:** Make sure to replace `C:\path	o\your\project\OpenPM.exe` with the actual absolute path to your compiled executable.

### 4. Managing the Task

*   **To view all scheduled tasks:**
    ```bash
    schtasks /query
    ```
*   **To view your specific task:**
    ```bash
    schtasks /query /tn "OpenPM Daily Check-in"
    ```
*   **To delete the task:**
    ```bash
    schtasks /delete /tn "OpenPM Daily Check-in"
    ```
*   **To run the task immediately for testing:**
    ```bash
    schtasks /run /tn "OpenPM Daily Check-in"
    ```
