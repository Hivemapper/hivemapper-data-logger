import math
import utm
import numpy as np
from numpy.linalg import inv


def make_F(dt):
    # F = State transition matrix (4x4)
    F = np.array([[1, dt, 0, 0], [0, 1, 0, 0], [0, 0, 1, dt], [0, 0, 0, 1]])
    return F


def make_G(dt):
    # G = Control matrix (4x2)
    G = np.array([[0.5 * dt * dt, 0], [dt, 0], [0, 0.5 * dt * dt], [0, dt]])
    return G


def make_H():
    # H = Measurement equation (2x4)
    H = np.array(
        [[1, 0, 0, 0], [0, 0, 1, 0]])  # Selects the position (x,y) state variables, since that is what we measure
    return H


def make_R(radial_uncertianty):  # (NmxNm), where Nm=number_of_measurements. This is just 2 measurements if using GP
    # R = Measurement covariance (2x2)
    R = np.array([[radial_uncertianty, 0], [0, radial_uncertianty]])
    return R


def make_K(H, P, R):
    # K = Kalman filter gain
    H_T = np.transpose(H)
    K1 = P @ H_T
    K2 = inv(H @ P @ H_T + R)
    K = K1 @ K2
    return K


def make_Q(G):
    # Process noise (4x4). Using method described in https://www.kalmanfilter.net/covextrap.html#withQ
    G_T = np.transpose(G)
    Q = G @ G_T
    return Q


def make_Uin(aR, aF, heading):
    pi = math.pi
    # print("aR=",aR, "  aF=", aF,"  heading=",heading)
    N1 = aF * math.cos((heading) * pi / 180.0)
    N2 = aR * math.sin((heading + 180) * pi / 180.0)
    E1 = aF * math.sin((heading) * pi / 180.0)
    E2 = aR * math.cos((heading) * pi / 180.0)
    N = N1 + N2
    E = E1 + E2
    return np.array([[E], [N]])


def update_state(X_in, K, H, z):  # z: 2x1, H: 2x4, X_in: 4x1, K: 4x2
    # print("X_in shape: ", np.shape(X_in),X_in)
    # print("K: shape: ", np.shape(K),K)
    # print("H: shape: ", np.shape(H),H)
    # print("z: shape: ", np.shape(z),z)
    # print("H @ X_in shape: ", np.shape(H @ X_in),H @ X_in)
    X = X_in + K @ (z - H @ X_in)
    return X  # 4x1


def update_covariance(P_in, K, H, R):
    K_T = np.transpose(K)
    I = np.identity(4)
    P11 = (I - K @ H)
    P11_T = np.transpose(P11)
    P1 = P11 @ P_in @ P11_T
    P2 = K @ R @ K_T
    P = P1 + P2
    return P


def predict_state(X_in, U_in, F, G):
    # X = F @ X_in
    print("term1 = ", F @ X_in)
    print("term2 = ", G @ U_in)
    X = F @ X_in + G @ U_in  # We use G and U_in since we assume a control input
    # print("F: \n",F)
    # print("G: \n",G)
    # print("U_in: \n",U_in)
    # print("X_in: \n",X_in)
    return X


def predict_covariance(P_in, F, Q):
    F_T = np.transpose(F)
    P = F @ P_in @ F_T + Q
    return P


def latlon_to_utm(lat, lon):
    utm_N = []
    utm_E = []
    for i in range(0, len(lat)):
        out = utm.from_latlon(lat[i], lon[i])
        utm_E.append(out[0])
        utm_N.append(out[1])
    return [utm_E, utm_N]


def find_heading(dE, dN):
    if dE == 0.0 and dN > 0.0:
        return 0.0
    if dE == 0.0 and dN < 0.0:
        return 180.0
    if dE > 0.0 and dN == 0.0:
        return 90.0
    if dE < 0.0 and dN == 0.0:
        return 270.0
    # For angles not along an axis ...
    tmp = math.atan(dE / dN) * (180.0 / math.pi)
    if dE > 0 and dN > 0:  # Q1
        return tmp
    if dE > 0 and dN < 0:  # Q2
        return 90.0 - tmp
    if dE < 0 and dN < 0:  # Q3
        return 180 + tmp
    if dE < 0 and dN > 0:  # Q4
        return 360 + tmp
    else:
        return None
