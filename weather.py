#!/usr/bin/env python3
import requests
import json
import sys
import time
from rich.console import Console
from rich.markdown import Markdown
from rich.live import Live
from rich.spinner import Spinner

def get_auth_token(console):
    """
    Gets authentication token from Jellyfaas API
    """
    auth_url = "https://api.jellyfaas.com/auth-service/v1/validate"
    headers = {
        'x-jf-apikey': '<apikey>'
    }
    
    try:
        response = requests.get(auth_url, headers=headers)
        response.raise_for_status()
        token_data = response.json()
        return token_data["token"]
    except requests.exceptions.RequestException as e:
        console.print(f"[bold red]Error getting authentication token: {e}[/bold red]")
        if hasattr(e, 'response') and e.response is not None:
            console.print(f"[bold red]Response status: {e.response.status_code}[/bold red]")
            console.print(f"[bold red]Response body: {e.response.text}[/bold red]")
        sys.exit(1)
    except (json.JSONDecodeError, KeyError) as e:
        console.print(f"[bold red]Error processing authentication response: {e}[/bold red]")
        sys.exit(1)

def get_weather_comparison(city1, city2, console):
    """
    Makes a request to the weather comparison service with the provided cities.
    Shows a spinner while waiting for the response.
    """

    # Get authentication token
    token = get_auth_token(console)

    url = "https://ai.jellyfaas.com/query-service/v1/function"
    
    headers = {
        'jfwt': token,
        'Content-Type': 'application/json'
    }
    
    payload = {
        "query": f"Compare the weather in {city1} and {city2} for the next three days, output this in a markdown table so it looks clean. Also give a section on any insights you have, or can make from the data, and give a fun intresting fact about both cities",
        "function": "weathercompare"
    }
    
    # Create a spinner
    spinner = Spinner('dots', text=f"Fetching weather comparison for {city1} and {city2}...")
    
    try:
        # Display the spinner while making the request
        with Live(spinner, console=console, refresh_per_second=10):
            response = requests.post(url, headers=headers, json=payload)
            response.raise_for_status()  # Raise an exception for 4XX/5XX responses
        return response.json()
    except requests.exceptions.RequestException as e:
        console.print(f"[bold red]Error making request: {e}[/bold red]")
        sys.exit(1)
    except json.JSONDecodeError:
        console.print("[bold red]Error: Could not decode JSON response[/bold red]")
        sys.exit(1)

def display_response(response_data, console):
    """
    Display the response with proper markdown rendering.
    """
    if "answer" in response_data and response_data["answer"]:
        # Convert the answer to Markdown
        markdown = Markdown(response_data["answer"])
        
        # Render the markdown
        console.print(markdown)
    else:
        console.print("[bold yellow]No weather data available in the response.[/bold yellow]")

def main():
    # Initialize Rich console
    console = Console()
    
    console.print("[bold blue]=== Weather Comparison Tool ===[/bold blue]")
    
    # Get city names from the user
    city1 = console.input("[green]Enter the first city name: [/green]")
    city2 = console.input("[green]Enter the second city name: [/green]")
    
    if not city1 or not city2:
        console.print("[bold red]Error: Both city names are required.[/bold red]")
        sys.exit(1)
    
    # Make the API request with spinner
    response_data = get_weather_comparison(city1, city2, console)
    
    # Display the results
    display_response(response_data, console)

if __name__ == "__main__":
    main()
