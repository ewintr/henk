# Henk - The AI assistant that lets you write your own code

An experimental AI tool designed to help and coach you, but that will not do your work for you.  Henk aims to be a collaborative partner in your coding process, not a replacement for your skills.

## Why?

- LLMs produce code that often requires adaptation. They have a tendency to overengineer things, introduce subtle bugs, pick the wrong names for variables, etc. Specially when generating large changes, it can be challenging to explain what exactly you want different. 
- If the code does not pass through your hands, you will not build a good mental map of the project. It is crucial for long-term maintainability that you _know_ what happens in the project.
- An assistant should complement your toolchain, not replace it. You have already learned to edit text efficiently, no need to ask Claude "Please name foo to bar" and hope for the best. Leave the text manipulation to your favorite editor/IDE, and let the assistant only worry about applying LLMs.

## What Henk will do

Currently in an early alpha stage, Henk can't do much yet. But the list below serves as a wish list for what could be implemented.

In the future:

- Henk can read your code, but cannot modify it, ensuring you remain in control.
- Henk will provide:
  - **Guidance:** Describe the feature you want to implement and Henk will help you break down complex features into smaller, manageable tasks, suggesting relevant algorithms or design patterns.
  - **Suggestions:**  For example, identifying opportunities to refactor code for improved readability or performance.
  - **Reviews:**  Highlighting potential bugs, security vulnerabilities, or code style issues.
  - **Snippets:**  Generating boilerplate code for common tasks, such as creating unit tests or API endpoints.
- Henk will be designed to be remote controlled, for easy integration with your favorite editor.
- Henk will work well with local LLMs
