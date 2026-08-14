const apiBase = "/v1/todos";

const state = {
  todos: [],
  nextCursor: null,
  editing: null,
};

const elements = {
  form: document.querySelector("#todoForm"),
  title: document.querySelector("#titleInput"),
  description: document.querySelector("#descriptionInput"),
  status: document.querySelector("#statusInput"),
  dueDate: document.querySelector("#dueDateInput"),
  statusFilter: document.querySelector("#statusFilter"),
  refresh: document.querySelector("#refreshButton"),
  cancel: document.querySelector("#cancelButton"),
  submit: document.querySelector("#submitButton"),
  list: document.querySelector("#todoList"),
  template: document.querySelector("#todoTemplate"),
  message: document.querySelector("#message"),
  summary: document.querySelector("#summary"),
  loadMore: document.querySelector("#loadMoreButton"),
};

elements.form.addEventListener("submit", async (event) => {
  event.preventDefault();
  await saveTodo();
});

elements.refresh.addEventListener("click", () => loadTodos());
elements.statusFilter.addEventListener("change", () => loadTodos());
elements.cancel.addEventListener("click", resetForm);
elements.loadMore.addEventListener("click", () => loadTodos(state.nextCursor));

loadTodos();

async function loadTodos(cursor = null) {
  setBusy(true);
  setMessage("");

  try {
    const params = new URLSearchParams({
      limit: "20",
      sort: "created_at_desc",
    });
    if (elements.statusFilter.value) {
      params.set("status", elements.statusFilter.value);
    }
    if (cursor) {
      params.set("cursor", cursor);
    }

    const response = await fetch(`${apiBase}?${params.toString()}`);
    const payload = await parseResponse(response);

    if (cursor) {
      state.todos = [...state.todos, ...payload.todos];
    } else {
      state.todos = payload.todos;
    }
    state.nextCursor = payload.next_cursor || null;
    renderTodos();
  } catch (error) {
    setMessage(error.message, true);
  } finally {
    setBusy(false);
  }
}

async function saveTodo() {
  setBusy(true);
  setMessage("");

  const payload = {
    title: elements.title.value,
    description: elements.description.value,
    status: elements.status.value,
    due_date: localDateTimeToISO(elements.dueDate.value),
  };

  try {
    let response;
    if (state.editing) {
      payload.version = state.editing.version;
      response = await fetch(`${apiBase}/${state.editing.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
    } else {
      response = await fetch(apiBase, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
    }

    await parseResponse(response);
    resetForm();
    await loadTodos();
  } catch (error) {
    setMessage(error.message, true);
  } finally {
    setBusy(false);
  }
}

async function deleteTodo(todo) {
  setBusy(true);
  setMessage("");

  try {
    const response = await fetch(`${apiBase}/${todo.id}`, { method: "DELETE" });
    if (!response.ok) {
      await parseResponse(response);
    }
    if (state.editing && state.editing.id === todo.id) {
      resetForm();
    }
    await loadTodos();
  } catch (error) {
    setMessage(error.message, true);
  } finally {
    setBusy(false);
  }
}

function renderTodos() {
  elements.list.replaceChildren();

  for (const todo of state.todos) {
    const item = elements.template.content.cloneNode(true);
    item.querySelector(".todo-title").textContent = todo.title;
    item.querySelector(".todo-description").textContent = todo.description || "";
    item.querySelector(".todo-meta").replaceChildren(
      pill(formatStatus(todo.status)),
      pill(`v${todo.version}`),
      pill(todo.due_date ? formatDate(todo.due_date) : "No due date"),
    );

    item.querySelector(".edit-button").addEventListener("click", () => {
      editTodo(todo);
    });
    item.querySelector(".delete-button").addEventListener("click", () => {
      deleteTodo(todo);
    });

    elements.list.appendChild(item);
  }

  elements.summary.textContent = `${state.todos.length} item${
    state.todos.length === 1 ? "" : "s"
  }`;
  elements.loadMore.hidden = !state.nextCursor;

  if (state.todos.length === 0) {
    setMessage("No TODOs");
  }
}

function editTodo(todo) {
  state.editing = todo;
  elements.title.value = todo.title;
  elements.description.value = todo.description || "";
  elements.status.value = todo.status;
  elements.dueDate.value = isoToLocalDateTime(todo.due_date);
  elements.submit.textContent = "Save";
  elements.cancel.hidden = false;
  elements.title.focus();
}

function resetForm() {
  state.editing = null;
  elements.form.reset();
  elements.status.value = "pending";
  elements.submit.textContent = "Add";
  elements.cancel.hidden = true;
}

function pill(text) {
  const element = document.createElement("span");
  element.textContent = text;
  return element;
}

async function parseResponse(response) {
  if (response.status === 204) {
    return null;
  }

  const payload = await response.json();
  if (!response.ok) {
    const error = payload.error;
    throw new Error(error ? error.message : `HTTP ${response.status}`);
  }

  return payload;
}

function setBusy(isBusy) {
  for (const element of [
    elements.submit,
    elements.refresh,
    elements.loadMore,
    elements.cancel,
  ]) {
    element.disabled = isBusy;
  }
}

function setMessage(text, isError = false) {
  elements.message.textContent = text;
  elements.message.classList.toggle("error", isError);
}

function localDateTimeToISO(value) {
  if (!value) {
    return null;
  }
  return new Date(value).toISOString();
}

function isoToLocalDateTime(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  const offsetMs = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

function formatDate(value) {
  return new Intl.DateTimeFormat("ja-JP", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatStatus(status) {
  return status.replace("_", " ");
}
