# OpenPM
OpenPM is a collection of technologies that creates a lightweight Project Manager to assist with solo developers or small groups.

The goal is to create a stack where the "PM" will reach out to the developer on a schedule, asking about the status of tasks to help improve accountability.

My current plan (as of the initial ideation of this project) is to use a locally-run LLM with a system prompt to work as a PM. This LLM will use RAG or an MCP server to retrieve data on a GitHub repository, where it will determine current tasks and decide what may need to be checked in on. It should also utilize either a daily notes system (so the developer can note their progress) or maintain conversation history for the latest reported status of the task. The LLM will be prompted on some cadence, possibly daily towards EOD during the week, to ask for updates and try to provide direction if any is needed. The response will be sent to the developer, possibly via webhook (so various systems such as e.g. a Discord bot could report it), which the developer can then respond to. Once all responses are finished, it will not respond until the next appropriate time (likely the next scheduled time).

Future improvements could include a custom scheduler, which allows for e.g. weekend work if prompted ("I'll have time this Saturday, can you check in on me then?").