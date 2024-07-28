import { useEffect, useState } from "react";
import PropTypes from "prop-types";
import axios from "axios";

const UserList = ({ role = 'user', users = [] }) => {
  const [localUsers, setLocalUsers] = useState(users); // Use local state for users
  const [error, setError] = useState("");
  const [companyId, setCompanyId] = useState(null);

  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const response = await axios.get("http://localhost:8080/get-user-info", { withCredentials: true });
        setCompanyId(response.data.companyId);
      } catch (error) {
        console.error("Error fetching user info:", error);
        setError("Failed to fetch user info");
      }
    };

    fetchUserInfo();
  }, []);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const url = role === 'companyAdmin'
          ? `http://localhost:8080/users?company_id=${companyId}` // Fetch users for the companyAdmin's company
          : "http://localhost:8080/users"; // Fetch all users for other roles
        
        const response = await axios.get(url, { withCredentials: true });
        setLocalUsers(response.data); // Update local state with fetched users
      } catch (error) {
        console.error("Error fetching users:", error);
        setError("Failed to fetch users: " + (error.response?.data || error.message));
      }
    };

    if (companyId !== null || role !== 'companyAdmin') {
      fetchUsers();
    }
  }, [role, companyId]);

  const handleDelete = async (userID) => {
    if (role === 'companyAdmin') {
      alert("You do not have permission to delete users.");
      return;
    }
    
    try {
      await axios.post("http://localhost:8080/delete-user", {
        user_id: userID,
      });
      alert("User deleted successfully");
      setLocalUsers(localUsers.filter((user) => user.id !== userID));
    } catch (error) {
      console.error("Error deleting user:", error);
      alert("Failed to delete user: " + (error.response?.data || error.message));
    }
  };

  return (
    <div className='user-list'>
      <div className='user-list-title'>
        <h2>User List</h2>
      </div>
      {error && <div style={{ color: "red" }}>{error}</div>}
      <div className='user-list-list'>
        <ul>
          {localUsers.map((user) => (
            <li key={user.id}>
              <strong>{user.username}</strong> - {user.email} - {user.role} - {user.company_name}
              {role !== 'companyAdmin' && (
                <button onClick={() => handleDelete(user.id)}>Delete</button>
              )}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

UserList.propTypes = {
  role: PropTypes.string, // Make role optional
  users: PropTypes.array // Make users optional
};

export default UserList;
