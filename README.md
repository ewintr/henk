# Henk - Local AI Code Assistant

**Note: ** Henk is currently in alpha. Examples and instructions on how to use it will be added shortly.

## Ideal end application

Henk is an AI assistant that helps you build software by mentoring and advising you, not by doing the work for you. The goal is not only to deliver good software, but to let the developer understand the inner workings of the project they are building. This is accomplished by letting the user make all the changes. Henk is given access to all parts of the project, to local documentation and to the internet, but does not have the power to change anything. In a sense Henk is a virtual pair programmer, the one with the navigator role. The user is the driver and controls the keyboard, mouse and is the one that applies the changes to the project.

Henk is composable with other tools, not a replacement for them. Henk does not contain an editor, a shell, etc. The user must be able to utilize the features of existing editors and other apps.

Henk strives to make the user programmer as independent as possible. That means, as mentioned, that the user must understand how the project works in such a way thet they can continue working on it without Henk. But it also means independent of cloud services. Henk must be able to work with different LLMs, both online and local.

## Flows

- Brainstorm on the architecture of new features
- Review and explain the current architecture
- Create a plan for implementing new feature
- Guiding the user step by step to follow that plan. Adjust the plan where necessary (new insights, user decided to change it)
- Help the user learn an unfamiliar language or tool that is used in the project

- Other types of assistant? Like for writing?

## Capabilities

- Read code
- Read local documentation
- Search the web for background information
- Generate snippets
- Monitor changes to the project

## UI

- Chat window
- Copy/paste - simple way of bringing in generate code snippets
- 'Remote control' - Sending a chat message from another program,

Perhaps later:
- Voice
- Realtime watching and interpreting the screen? 

## Possible Components

- Internal ToDo-list, like Claude?
- LSP-client, for better code understanding?

