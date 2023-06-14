from matplotlib import pyplot as plt
# Plot the output results
plt.scatter(sensorS_data[1],sensorS_data[0])
plt.show()
plt.scatter(Y_pos[0:],X_pos[0:])
plt.show()
# print(Y_pos[1:])
# print(X_pos[1:])
# print(sensorS_data[2])
# ################################################################################
# # Display ACCEL
# import altair as alt
# alt.Chart(accel_df).mark_point().encode(
#   x='time',
#   y='a2',
# ).interactive()