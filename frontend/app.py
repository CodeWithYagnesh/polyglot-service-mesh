from flask import Flask, render_template, request, redirect, session, url_for, flash
import os
import requests

app = Flask(__name__)
app.secret_key = os.environ.get('SECRET_KEY', 'your_secret_key')
AUTH_URL = os.environ.get('AUTH_URL', "http://localhost:8081")
BY_WHO_URL = os.environ.get('BY_WHO_URL', "http://localhost:8082")
CATEGORY_URL = os.environ.get('CATEGORY_URL', "http://localhost:8083")
ENTRY_URL = os.environ.get('ENTRY_URL', "http://localhost:8084")
PORT = os.environ.get('PORT', "3000")

@app.route('/')
def home():
    return redirect(url_for('login'))

@app.route('/register', methods=['GET', 'POST'])
def register():
    if request.method == 'POST':
        data = {
            "email": request.form['email'],
            "username": request.form['username'],
            "password": request.form['password']
        }
        res = requests.post(f"{AUTH_URL}/user/register", json=data)
        if res.ok:
            flash('Registration successful. Please login.', 'success')
            return redirect(url_for('login'))
    return render_template('register.html')

@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        data = {
            "email": request.form['email'],
            "password": request.form['password']
        }
        res = requests.post(f"{AUTH_URL}/user/login", json=data)
        if res.ok:
            session['userId'] = res.json()['userId']
            return redirect(url_for('dashboard'))
        else:
            flash('Login failed!', 'danger')
    return render_template('login.html')

@app.route('/dashboard')
def dashboard():
    if 'userId' not in session:
        return redirect(url_for('login'))
    return render_template('dashboard.html')

@app.route('/dashboard/bywho/add', methods=['GET', 'POST'])
def add_bywho():
    if request.method == 'POST':
        data = {
            "userId": session['userId'],
            "name": request.form['name'],
            "description": request.form['description']
        }
        res = requests.post(f"{BY_WHO_URL}/bywho/add", json=data)
        if res.ok:
            flash('ByWho added successfully.', 'success')
        else:
            flash('Failed to add ByWho.', 'danger')
        return redirect(url_for('dashboard'))
    return render_template('add_bywho.html')

@app.route('/dashboard/category/add', methods=['GET', 'POST'])
def add_category():
    if request.method == 'POST':
        data = {
            "userId": session['userId'],
            "name": request.form['name'],
            "description": request.form['description']
        }
        requests.post(f"{CATEGORY_URL}/category/add", json=data)
        return redirect(url_for('dashboard'))
    return render_template('add_category.html')

@app.route('/dashboard/entry/add', methods=['GET', 'POST'])
def add_entry():
    if request.method == 'POST':
        data = {
            "userId": session['userId'],
            "transaction_type": request.form['transaction_type'],
            "owe": bool(request.form.get('owe')),
            "date": request.form['date'],
            "reason": request.form['reason'],
            "by_who": request.form['by_who'],
            "category": request.form['category'],
            "amount": request.form['amount'],
            "oweList": []  # Add support later if needed
        }
        requests.post(f"{ENTRY_URL}/entry/add", json=data)
        return redirect(url_for('list_entries'))
    return render_template('add_entry.html')

@app.route('/dashboard/entries')
def list_entries():
    user_id = session.get('userId')
    if not user_id:
        return redirect(url_for('login'))

    # Example date range - last 30 days
    from datetime import date, timedelta
    date_to = date.today()
    date_from = date_to - timedelta(days=30)

    res = requests.get(f"{ENTRY_URL}/entry/list?userId={user_id}&dateFrom={date_from}&dateTo={date_to}")
    entries = res.json() if res.ok else []
    return render_template('entries.html', entries=entries)

@app.route('/logout')
def logout():
    session.clear()
    return redirect(url_for('login'))

if __name__ == '__main__':
    if PORT != None:
        PORT=3000
    app.run(host="0.0.0.0",port=int(PORT))
