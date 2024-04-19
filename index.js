document.getElementById('expression-form').addEventListener('submit', async (event) => {
    event.preventDefault();

    const expression = document.getElementById('expression').value;
    const response = await fetch('/add?expression=' + encodeURIComponent(expression));
    const id = await response.text();

    addExpressionToTable(id, expression, 'waiting', '');

    const resultResponse = await fetch('/expression?id=' + id);
    const resultData = await resultResponse.json();

    updateExpressionInTable(resultData);
});

function addExpressionToTable(id, expression, status, result) {
    const tbody = document.getElementById('results-table').getElementsByTagName('tbody')[0];
    const row = tbody.insertRow();

    row.insertCell().textContent = id;
    row.insertCell().textContent = expression;
    row.insertCell().textContent = status;
    row.insertCell().textContent = result;
}

function updateExpressionInTable(expression) {
    const rows = document.getElementById('results-table').getElementsByTagName('tbody')[0].rows;

    for (let i = 0; i < rows.length; i++) {
        const row = rows[i];
        const cells = row.cells;

        if (cells[0].textContent === expression.id) {
            cells[1].textContent = expression.expression;
            cells[2].textContent = expression.status;
            cells[3].textContent = expression.result;
            break;
        }
    }
}
