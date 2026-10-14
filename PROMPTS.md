# Prompt 1. Plan the application

I want to build a Book Tracker web application using plain HTML, CSS, and vanilla JavaScript.

The application should allow users to:
- Add a landing page and add option to continue as guest
- Add books with a title, author, and reading status.
- Display books with the newest additions first.
- Search books by title or author.
- Change a book's reading status.
- Delete books.
- Save books so they remain available after refreshing the browser.

Before writing any code, explain the proposed application structure, list the files you will create, describe how each file will be used, and explain how the data will be stored.

Do not write the code yet. Keep the solution beginner-friendly and avoid unnecessary dependencies.

# Prompt 2. Create the project files

Based on the plan, create the initial project files:

- index.html
- style.css
- app.js

Use semantic HTML, clean CSS, and vanilla JavaScript.

For now, create only the basic HTML page structure, link the CSS and JavaScript files correctly, and add a simple heading saying "Book Tracker".

Do not implement the other features yet. Explain how I can open and run the application locally.

# Prompt 3. Design the user interface

Improve the Book Tracker interface using HTML and CSS.

Create a clean, modern, responsive layout with:
- A header containing the application name.
- A form for entering a book title and author.
- A dropdown with the statuses To Read, Reading, and Finished.
- An Add Book button.
- A section for displaying the book list.
- A search input above the list.

Use a readable font, accessible color contrast, clear spacing, and simple status labels. Make it usable on both desktop and mobile screens.

Only implement the visual layout at this stage. Do not add application logic or external libraries. Keep all styling in style.css.

# Prompt 4. Add a book

Implement the Add Book feature in app.js.

When the user submits the form:
1. Read the title, author, and selected reading status.
2. Validate that the title and author are not empty or whitespace-only.
3. Show a clear validation message near the relevant input if either field is invalid.
4. If the inputs are valid, create a book object with a unique ID, title, author, status, and creation timestamp.
5. Add the new book to the top of the list.
6. Clear the form after a successful addition.

Prevent the form from reloading the page when submitted. Use safe DOM methods and textContent when displaying user-provided text.

Do not implement localStorage, search, or status editing yet. Explain the changes you made.

# Prompt 5. Display the book list

Implement the book-list rendering logic in app.js.

Display each book with:
- Its title.
- Its author.
- Its current reading status.
- A control for changing its status.
- A Delete button.

Display the newest book first. If there are no books, show a friendly empty-state message.

Separate the rendering logic into a reusable function so it can be called whenever the book data changes.

Do not introduce a framework or additional dependencies. Test the rendering with several sample books, then explain how to remove any temporary test data.

# Prompt 6. Change the reading status

Implement the ability to change a book's reading status.

Each book should have a dropdown with three options:
- To Read
- Reading
- Finished

When a user changes the dropdown, update the corresponding book object using its unique ID and re-render the list.

Ensure that changing one book's status does not accidentally change another book's status.

Handle invalid or unexpected status values safely. Do not implement persistent storage yet.

# Prompt 7. Delete a book

Implement the Delete feature for each book.

When the user clicks Delete:
1. Identify the selected book using its unique ID.
2. Remove only that book from the book collection.
3. Re-render the list.
4. Show the empty-state message if the last book is deleted.

If appropriate, add a simple confirmation before deletion, but keep the interaction straightforward and accessible.

Do not implement localStorage yet. Explain how you ensure the correct book is deleted.

# Prompt 8. Save books with localStorage

Implement persistent storage using browser localStorage.

Requirements:
- Save the complete book collection whenever a book is added, updated, or deleted.
- Load saved books when the application starts.
- Preserve each book's ID, title, author, status, and creation timestamp.
- Keep the newest-first ordering.
- Handle missing, invalid, or malformed saved JSON without crashing the application.
- If storage is unavailable or saving fails, display an appropriate message rather than silently claiming the data was saved.

Keep the storage logic in reusable functions. Do not add a backend or external library.

Explain how I can verify that the data survives a browser refresh.

# Prompt 9. Implement search

Implement live search for the Book Tracker.

When the user types into the search input, display only books whose title or author contains the search query.

Requirements:
- Ignore differences between uppercase and lowercase letters.
- Match partial words.
- Trim unnecessary whitespace from the query.
- Update the displayed results as the user types.
- Preserve the original book collection while searching.
- Show a helpful message when no books match.
- Restore the complete list when the search input is cleared.

Make sure search works together with adding, deleting, and changing book statuses.

# Prompt 10. Improve the user experience

Review the Book Tracker interface and improve its usability without adding unnecessary complexity.

Please:
- Add a visible count of all books.
- Show counts for To Read, Reading, and Finished.
- Add a clear-search control when a query is entered.
- Ensure validation messages are understandable and accessible.
- Make buttons and dropdowns keyboard-accessible.
- Improve the empty state and no-search-results state.
- Check that long book titles and author names do not break the layout.
- Preserve all existing features and data.

Make only changes that improve usability. Explain which files were modified and how I can test each improvement.

# Prompt 11. Run a structured test review

Review the current Book Tracker implementation as a software tester.

Create a test checklist covering:
- Adding valid books.
- Empty and whitespace-only inputs.
- Very long titles and author names.
- Changing statuses independently.
- Deleting one book and deleting the last book.
- Searching with uppercase and lowercase letters.
- Searching for partial matches.
- Clearing a search.
- Refreshing the page after changing data.
- Invalid or corrupted localStorage data.
- Mobile layout and keyboard accessibility.

For each test, describe the steps, expected result, and actual result if you can verify it.

Do not claim to have executed tests unless you actually ran them. Identify any code-level issues and recommend the smallest fixes first.

# Prompt 12. Test edge cases

Perform a focused edge-case review of the Book Tracker.

Inspect how the application handles:
1. Titles or author names containing only spaces.
2. Very long strings.
3. Text containing HTML-like input such as <script>alert(1)</script>.
4. Invalid reading-status values in stored data.
5. Corrupted localStorage JSON.
6. Two books with similar titles but different IDs.
7. Deleting a book while a search is active.
8. Changing a status while a search is active.

Identify reproducible bugs and explain why they occur. Fix only the confirmed problems, preserve existing functionality, and provide clear manual verification steps for each fix.

# Prompt 13. Reproduce and fix a bug

Help me investigate a bug in my Book Tracker.

First, inspect the relevant code and ask me to reproduce the suspected issue if the evidence is insufficient.

For each confirmed bug:
1. Describe the steps needed to reproduce it.
2. State the expected behavior.
3. State the actual behavior I observed.
4. Explain the root cause based on the code.
5. Implement the smallest appropriate fix.
6. Explain how to verify the fix and check for regressions.

Do not invent test results or claim that a bug is fixed without providing a way to verify it.

Here is the bug or unexpected behavior I observed:

[Replace this section with the actual bug and relevant code.]

# Prompt 14. Final regression test

Review the complete Book Tracker after the recent changes.

Verify that:
- Adding a book works.
- Required fields are validated.
- Newest-first ordering is preserved.
- Status changes work.
- Deletion works.
- Search works with partial and case-insensitive matches.
- Data persists after refresh.
- Corrupted saved data does not crash the application.
- Existing features still work after bug fixes.
- The responsive layout remains usable.

Inspect for duplicated code, incorrect event handling, unsafe rendering, and unnecessary complexity.

Do not rewrite the application from scratch. Make only necessary corrections and provide a final checklist showing which items were actually tested and which still require manual verification.


Keep the documentation beginner-friendly and accurate to the implementation. Do not claim that the application uses a backend, authentication, automated tests, or deployment if those features have not been implemented.

# Prompt 15. Write the README   
- Created README.md with overview, features, technologies, file structure, run instructions, localStorage explanation, testing instructions, known limitations, and future improvements.