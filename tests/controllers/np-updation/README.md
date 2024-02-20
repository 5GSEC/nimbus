# Description

This test ensures the persistant nature of nimbus-policy.

# Expected Behaviour

On updation of NimbusPolicy the policy by modifying the value of the key in .spec.selector.matchLablels.app the policy should get reverted back to its original state.