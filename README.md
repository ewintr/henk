# Henk - The AI assistant that lets you write your own code

An experimental AI tool designed to help and coach you, but that will not do your work for you without your involvement.  Henk aims to be a collaborative partner in your coding process, not a replacement for your skills.

## The Problem

- LLMs produce code that often requires adaptation. They have a tendency to overengineer things, introduce subtle bugs, pick the wrong names for variables, etc. Specially when generating large changes, it can be challenging to explain in a prompt what exactly you want different.
- LLMs quickly become unreliable when you venture off the beaten path, which limits the range of solutions that can suggest. This leads to suboptimal code that is difficult to correct afterwards.
- If the code does not pass through your hands, you will not build a good mental map of the project. It is crucial for long-term maintainability that you _know_ what happens in the project.
- An assistant should complement your toolchain, not replace it. You have already learned to edit text efficiently. No need to ask Claude "Please rename foo to bar" and hope for the best. Leave the text manipulation to your favorite editor/IDE, and let the assistant only worry about applying LLMs.

## The Solution

A code assistant that supports you well working, but that lets you hold the reigns. Like peer programming with a virtual colleague, but you are the one who is driving.

## What Henk will do

Key Principle: Henk will never directly modify your code.

Henk will focus on augmenting your workflow with:

* **Code Understanding & Analysis:**
  * **Guidance:** Describe the feature you want to implement and Henk will help you break down complex features into smaller, manageable tasks, suggesting relevant algorithms or design patterns.
  * **Reviews:** Highlighting potential bugs, security vulnerabilities, or code style issues.
* **Code Generation & Assistance:**
  * **Suggestions:** Identifying opportunities to refactor code for improved readability or performance.
  * **Snippets:** Generating boilerplate code for common tasks, such as creating unit tests or API endpoints.

Henk will be designed for seamless integration with your existing tools:

* **Remote Control:** Henk will integrate with your editor via a plugin or API, allowing you to access its features without leaving your development environment.
* **Local LLM Support:** **Privacy and Control: Henk is designed to work seamlessly with local LLMs, giving you complete control over your data and costs.**

