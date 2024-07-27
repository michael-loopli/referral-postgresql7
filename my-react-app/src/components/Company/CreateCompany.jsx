import { useState } from "react";
import axios from "axios";
import "../styles.css";
import Companies from "./Companies";

const CreateCompany = () => {
  const [company, setCompany] = useState({
    name: "",
  });

  const handleChange = (e) => {
    setCompany({ ...company, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await axios.post("http://localhost:8080/create-company", company);
      alert("Company created successfully");
      setCompany({
        name: "",
      });
    } catch (error) {
      console.error("Error creating company:", error);
      alert("Failed to create company");
    }
  };

  return (
    <>
      <div>
        <h1>Create Company</h1>
        <form onSubmit={handleSubmit}>
          <label className='create-company-heading'>New Company Name:</label>
          <input
            type='text'
            name='name'
            value={company.name}
            onChange={handleChange}
            required
          />
          <br />
          <button type='submit'>Create Company</button>
        </form>
      </div>
      <div>
        <Companies />
      </div>
    </>
  );
};

export default CreateCompany;
