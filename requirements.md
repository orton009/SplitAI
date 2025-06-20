## Features 

1. Create new User with username and email 
2. Create Group with groupname and administrator 
3. Add Users to Group 
4. Create Expense inside a Group 
5. Create Expense with User outside of a Group
6. Export exepnses to CSV format
7. Simplify split across multiple groups and local user expenses 

## Expense Types 

1. Unit Split - 5 rs to u1, 10 rs to u2
2. Percentage Split - 10% to u1, 90% to u2
3. Share Split - 1/3 to u1, 2/3 to u2
4. Owed to Splits - Single person paid, multiple people paid
5. Equal Split - equally to all people

## User Interface 

1. Command Line interface
2. HTTP APIs
3. RPC

## Storage Interface

1. Local In memory Storage
2. Persistent File Storage
3. Relational Database
4. KV DB/REDIS storage


## Expense Metrics Overlook

UserBorrow : how much user owes in that specific context
userOwed : how much user borrowed in that specific context
userLiable: userBorrow or userOwed

context: current outstanding, only unsettled expenses.
1. Add current userLiability in specific group expense. 
2. show current userLiability in specific group
3. show totalExpense of the group (current outstanding, only unsettled expenses)
4. show totalExpense of group by time period (month range, date range, year range)
5. show current userLiability across all groups, overall owed/borrowed


### NEXT TO PICK UP
1. Add leave group logic, what if admin leaves ? new admin ? 
2. Add remove friend on frontend
3. Add show more on scrollable paginated api calls on frontend, test paginated calls

1. Add DB Transaction Locking in places
2. Add go routines for parallel processing (heavy calculations)
3. Remove any business/data Logic from storage layer apart from transformation required for storage
4. Remove any business logic from service, it should only have extra logic required for transformation/combination/db level validations. 