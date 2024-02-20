# Description

This test ensures the creation of nimbus-policy and assert the status subresources of SecurityIntent and SecurityIntentBinding.

# Expected Behaviour

Upon creation of SecurityIntent and SecurityIntentBinding the NimbusPolicy should get created. Furthermore the SecurityIntent subresource should contain the id and action field and SecurityIntentBinding status subresource reflects the number and names of bound intents, along with the generated NimbusPolicy name.
