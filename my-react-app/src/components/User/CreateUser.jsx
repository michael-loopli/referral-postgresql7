import { useState, useEffect } from "react";
import axios from "axios";
import PropTypes from "prop-types";
import UserList from "./UserList";
import "./CreateUser.css";

const CreateUser = ({ companies, refreshData, role }) => {
  const [newUser, setNewUser] = useState({
    email: "",
    username: "",
    password: "",
    role: "user", // Default role is 'user'
    company_id: "",
  });
  const [currentCompany, setCurrentCompany] = useState(""); // Store the current company for 'companyAdmin'
  const [users, setUsers] = useState([]); // List of users for 'companyAdmin'
  const [error, setError] = useState(null);

  // Initialize company data and users for 'companyAdmin'
  useEffect(() => {
    const fetchData = async () => {
      if (role === "companyAdmin") {
        try {
          // Fetch company info
          const response = await axios.get(
            "http://localhost:8080/get-user-info",
            { withCredentials: true }
          );
          const { companyId, companyName } = response.data;
          setCurrentCompany(companyName);
          setNewUser((prevData) => ({
            ...prevData,
            company_id: companyId,
          }));

          // Fetch users for the company
          const usersResponse = await axios.get(
            `http://localhost:8080/users?company_id=${companyId}`,
            { withCredentials: true }
          );
          setUsers(usersResponse.data);
        } catch (error) {
          console.error("Error fetching company info or users:", error);
          setError("Failed to fetch company info or users");
        }
      }
    };

    fetchData();
  }, [role]);

  // Handle user creation
  const handleCreateUser = async (e) => {
    e.preventDefault();
    try {
      const userToCreate = {
        ...newUser,
        company_id: parseInt(newUser.company_id),
      };
      await axios.post("http://localhost:8080/create-user", userToCreate);
      setNewUser({
        email: "",
        username: "",
        password: "",
        role: "user",
        company_id: "",
      });
      refreshData(); // Refresh data after creating user
      alert("User created successfully");
    } catch (error) {
      console.error("Error creating user:", error);
      setError("Failed to create user");
    }
  };

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <>
      <div className='user-container'>
        <section>
          <div className='create-user'>
            <h2>Create New User</h2>
            <form onSubmit={handleCreateUser}>
              <div className='user-input'>
                <label htmlFor='user-username'>Username:</label>
                <input
                  type='text'
                  id='user-username'
                  value={newUser.username}
                  onChange={(e) =>
                    setNewUser({ ...newUser, username: e.target.value })
                  }
                  required
                />
              </div>
              <br />
              <div className='user-input'>
                <label htmlFor='user-email'>Email:</label>
                <input
                  type='email'
                  id='user-email'
                  value={newUser.email}
                  onChange={(e) =>
                    setNewUser({ ...newUser, email: e.target.value })
                  }
                  required
                />
              </div>
              <br />
              <div className='user-input'>
                <label htmlFor='user-password'>Password:</label>
                <input
                  type='password'
                  id='user-password'
                  value={newUser.password}
                  onChange={(e) =>
                    setNewUser({ ...newUser, password: e.target.value })
                  }
                  required
                />
              </div>
              <br />
              <div className='user-input'>
                <label htmlFor='user-role'>Role:</label>
                <select
                  id='user-role'
                  value={newUser.role}
                  onChange={(e) =>
                    setNewUser({ ...newUser, role: e.target.value })
                  }
                  required>
                  {role === "companyAdmin" ? (
                    <option value='user'>Standard User</option>
                  ) : (
                    <>
                      <option value='platformAdmin'>Platform Admin</option>
                      <option value='companyAdmin'>Company Admin</option>
                      <option value='user'>Standard User</option>
                    </>
                  )}
                </select>
              </div>
              <br />

              <div>
                <label htmlFor='user-company'>Company:</label>
                {role === "companyAdmin" ? (
                  <input
                    type='text'
                    id='user-company'
                    value={`Company ID: ${newUser.company_id} (${currentCompany})`}
                    readOnly
                  />
                ) : (
                  <select
                    id='user-company'
                    value={newUser.company_id}
                    onChange={(e) =>
                      setNewUser({ ...newUser, company_id: e.target.value })
                    }
                    required>
                    <option value=''>Select a company</option>
                    {companies.length > 0 ? (
                      companies.map((company) => (
                        <option key={company.id} value={company.id}>
                          {company.name}
                        </option>
                      ))
                    ) : (
                      <option value=''>No companies available</option>
                    )}
                  </select>
                )}
                <br />
              </div>

              <div className='user-submit-container'>
                <button className='user-submit' type='submit'>
                  Create User
                </button>
              </div>
            </form>
          </div>
        </section>

        {role === "companyAdmin" && (
          <>
            <div className='user-list-cu'>
              <h3>Users in Your Company</h3>
              <UserList role={role} users={users} />
            </div>
          </>
        )}

        {(role === "superAdmin" || role === "platformAdmin") && (
          <div className='user-list-cu'>
            <UserList role={role} />
          </div>
        )}
      </div>
    </>
  );
};

CreateUser.propTypes = {
  companies: PropTypes.array.isRequired,
  refreshData: PropTypes.func.isRequired,
  role: PropTypes.string.isRequired,
};

export default CreateUser;
