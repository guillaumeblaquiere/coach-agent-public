#!/bin/bash

PROJECT_ID=<YOUR_PROJECT_ID>
go run .

#Activate the firestore API
gcloud services enable firestore.googleapis.com --project=${PROJECT_ID}

#Create the firestore database
gcloud firestore databases create --project=${PROJECT_ID}  --location=europe-west1


# Categories
# List Categories
curl http://localhost:8080/api/v1/categories

# Get a Specific Category (replace {categoryId})
# Example: curl http://localhost:8080/api/v1/categories/back
#  You'll need to know a valid category ID for this to work without modification.
#  To get valid IDs first, run the "List Categories" command above and examine the output.
curl http://localhost:8080/api/v1/categories/back


# Drills
# List All Drills
curl http://localhost:8080/api/v1/drills

# Get a Specific Drill (replace {drillId})
# Example:  curl http://localhost:8080/api/v1/drills/HipStretchedLeft
# To get valid drill IDs, you can run the "List Drills" command and inspect the output.
#curl http://localhost:8080/api/v1/drills/{drillId}  # Placeholder: Uncomment and replace!


# Training Plan Templates
# List Available Templates (currently only "default" is supported, the command is provided directly.)
curl http://localhost:8080/api/v1/plan-templates

# Get the Default Template (only "default" is currently supported, the command is provided directly.)
curl http://localhost:8080/api/v1/plan-templates/default


# Daily Training Plans

# Initiate a Daily Plan (Today, Default Template)  - No body needed if using defaults.
curl -X POST http://localhost:8080/api/v1/daily-plans/initiate

# Initiate a Daily Plan (Specific Date, Default Template)  (Replace YYYY-MM-DD)
# Example: curl -X POST -H "Content-Type: application/json" -d '{"date": "2024-07-29"}' http://localhost:8080/api/v1/daily-plans/initiate
# curl -X POST -H "Content-Type: application/json" -d '{"date": "YYYY-MM-DD"}' http://localhost:8080/api/v1/daily-plans/initiate # Placeholder: Uncomment and replace!

# Get Daily Plan (Today)
curl http://localhost:8080/api/v1/daily-plans/today

# Get Daily Plan (Specific Date) (Replace YYYY-MM-DD)
# Example: curl http://localhost:8080/api/v1/daily-plans/2024-07-29
#curl http://localhost:8080/api/v1/daily-plans/YYYY-MM-DD  # Placeholder: Uncomment and replace!

# Update Today's Daily Plan
# Get the plan first, modify the JSON, then PUT the updated version.
# Note: The update requires you to send the entire DailyTrainingPlan object back.
# Here's a multi-step approach
#
# 1. Get the current plan (today's) and store it in a variable:
# today_plan=$(curl -s http://localhost:8080/api/v1/daily-plans/today)
#
# 2.  Modify $today_plan's JSON.
#     Since `jq` isn't always available, this example does a minimal manual change within the script.
#     For more complex modifications, `jq` or a more robust JSON manipulation tool is recommended.
#
#     Here's how to add (or modify) the "Abs" drill in the "back" category to have a repetition value of 5:
#     updated_plan=$(echo "$today_plan" |
#          sed 's/"Abs".*"repetition": [0-9]*/"Abs", "repetition": 5/'
#     )
#
#     Important: Make sure there exists an "Abs" exercise id in a "back" category, and the "repetition" field exists!
#
#     For more complex edits with proper JSON,  consider using `jq`:
#     Example using jq (assuming you have it installed):
#     updated_plan=$(echo "$today_plan" | jq '.categories.back.drills += [{ "id": "Abs", "name": "Etirement abdominaux", "description": "Position au allongé face au sol, bras tendu, tête vers le haut", "categoryId": "back", "targetRepetition": 3, "notes": "", "repetition": 5, "createdAt": "2024-07-28T13:38:15.106285Z", "updatedAt": "2024-07-28T13:38:15.106285Z" }]')
#
# 3. Send the updated plan back to the server:
curl -X PUT -H "Content-Type: application/json" -d "@updateDrill.json" http://localhost:8080/api/v1/daily-plans/today
curl -X PUT -H "Content-Type: application/json" -d "@updateDrill.json" http://coach-backend-513069150666.europe-west1.run.app/api/v1/daily-plans/today

# Health Check
curl http://localhost:8080/health