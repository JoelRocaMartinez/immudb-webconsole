#!/usr/bin/env python3

from immudb.client import ImmudbClient

ic=ImmudbClient()
ic.login("immudb","immudb")

key="a_very_important_key".encode('utf8')
value="a_very_important_value".encode('utf8')

# let's insert the value in the DB and check
# if it was correctly inserted
response=ic.verifiedSet(key,value)

# here response is a structure holding many informations
# about the merkele tree, but the most important is that 
# the insert was correctly verified
assert response.verified==True

print("Key inserted (and verified) with index",response.index)

#reads back the value
readback=ic.verifiedGet(key)

# in the readback we also have the index and the verified field
assert response.verified==True
print("The value is",readback.value,"at transaction",response.id,"with timestamp",readback.timestamp)

